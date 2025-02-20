package mapper

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// metadata tag values
const (
	metaTagGeneration = "generation"
	metaTagExpiration = "expiration"
	metaTagNamespace  = "namespace"
	metaTagSetName    = "set_name"
	metaTagDigest     = "digest"
	metaTagUserKey    = "user_key"
)

const (
	mapperTag         = "aero"
	tagValueMeta      = "meta"
	tagValueOmit      = "omit"
	tagValueOmitempty = "omitempty"

	timeType = "time.Time"
)

var (
	reflectZeroValue = reflect.Value{}
)

// tag represents the `aero` tag to mark fields and specify their mapping.
type tag struct {
	// meta indicates that the field should be mapped to Aerospike record metadata (namespace,
	// set name, key, generation, expiration).
	meta bool
	// omit indicates that the field should be omitted entirely from the Aerospike record.
	omit bool
	// omitempty indicates that the field should be omitted from the Aerospike record if it is
	// empty (zero value).
	omitempty bool
	// name is the bin name to use for the field.
	name string
}

// Record is the Aerospike record representation produced by the Encode operation.
// It contains the key, metadata, and bin data.
type Record struct {
	// Key contains record's key details (namespace, set name, digest).
	Key
	// KeyValue contains record's user key.
	KeyValue
	// Metadata contains record's generation and expiration details.
	Metadata
	// Bins contains record's bin data (name-value pairs).
	Bins map[string]any
}

// Encode encodes v into a Record.
// v must be a struct or struct pointer with fields tagged using the `aero` tag
// to specify how they should be mapped to the record.
func Encode(v any) (*Record, error) {
	// initialize the return record value
	record := &Record{
		Bins: make(map[string]any),
	}
	// call the recursive encode function
	return encode(v, record)
}

// encode recursively encodes v and returns the encoded record.
//
//nolint:funlen
func encode(v any, record *Record) (*Record, error) {
	sourceValue, err := structValue(v)
	if err != nil {
		return nil, err
	}

	sourceType := sourceValue.Type()
	for i := 0; i < sourceType.NumField(); i++ {
		fieldValue := fieldValueDeref(sourceValue, i)
		if fieldValue.Kind() == reflect.Struct {
			_, err := encode(fieldValue, record)
			if err != nil {
				return nil, err
			}
			continue
		}

		aeroTag := sourceType.Field(i).Tag.Get(mapperTag)
		if aeroTag == "" {
			continue
		}

		tag, err := parseTag(aeroTag)
		if err != nil {
			return nil, err
		}

		if tag.meta {
			switch tag.name {
			case metaTagGeneration:
				err := setMetadata(fieldValue,
					reflect.ValueOf(&record.Metadata).Elem().FieldByName("Generation"), tag.name)
				if err != nil {
					return nil, err
				}
			case metaTagExpiration:
				err := setMetadata(fieldValue,
					reflect.ValueOf(&record.Metadata).Elem().FieldByName("Expiration"), tag.name)
				if err != nil {
					return nil, err
				}
			case metaTagNamespace:
				err := setMetadata(fieldValue,
					reflect.ValueOf(&record.Key).Elem().FieldByName("Namespace"), tag.name)
				if err != nil {
					return nil, err
				}
			case metaTagSetName:
				err := setMetadata(fieldValue,
					reflect.ValueOf(&record.Key).Elem().FieldByName("SetName"), tag.name)
				if err != nil {
					return nil, err
				}
			case metaTagDigest:
				err := setMetadata(fieldValue,
					reflect.ValueOf(&record.Key).Elem().FieldByName("Digest"), tag.name)
				if err != nil {
					return nil, err
				}
			case metaTagUserKey:
				err := setMetadata(fieldValue,
					reflect.ValueOf(&record.KeyValue).Elem().FieldByName("UserKey"), tag.name)
				if err != nil {
					return nil, err
				}
			}
		} else {
			// handle omit and omitempty tags
			empty := isEmptyValue(fieldValue)
			if tag.omit || (tag.omitempty && empty) {
				continue
			}
			binName := tag.name
			if binName == "" {
				// binName = sourceType.Field(i).Name
				continue
			}
			if empty {
				record.Bins[binName] = reflect.Zero(sourceType.Field(i).Type).Interface()
			} else {
				record.Bins[binName] = fieldValue.Interface()
			}
		}
	}

	return record, nil
}

// setMetadata is a helper function to set record metadata fields.
func setMetadata(field reflect.Value, recordField reflect.Value, tagName string) error {
	if isEmptyValue(field) {
		return nil
	}

	if field.Type() != recordField.Type() {
		return fmt.Errorf("type mismatch for tag '%s': source %s, record %s", tagName,
			field.Type(), recordField.Type())
	}

	if recordField.CanSet() {
		recordField.Set(field)
		return nil
	}

	return errors.New("cannot set record field")
}

// Decode decodes an aerospike record or a record containing struct into v.
func Decode(record, v any) error {
	_, inner := record.(reflect.Value)
	recordValue, err := structValue(record)
	if err != nil {
		return err
	}

	var isRecord bool
	recordType := recordValue.Type()
	for i := 0; i < recordType.NumField(); i++ {
		fieldValue := fieldValueDeref(recordValue, i)

		fieldName := recordType.Field(i).Name
		switch {
		case fieldValue.Kind() == reflect.Struct && fieldName == "BatchRecord":
			isRecord = true
			if err := Decode(fieldValue, v); err != nil {
				return err
			}
		case fieldValue.Kind() == reflect.Struct && fieldName == "Record":
			isRecord = true
			if err := decodeRecord(fieldValue, v); err != nil {
				return err
			}
			if err := Decode(fieldValue, v); err != nil {
				return err
			}
		case fieldValue.Kind() == reflect.Struct && fieldName == "Key":
			if err := decodeKey(fieldValue, v); err != nil {
				return err
			}
		case fieldValue.Kind() == reflect.Map && fieldName == "Bins":
			isRecord = true
			if err := decodeBins(fieldValue, v); err != nil {
				return err
			}
		case !inner && fieldValue.Kind() == reflect.Uint32 &&
			(fieldName == "Generation" || fieldName == "Expiration"):
			if err := decodeRecord(recordValue, v); err != nil {
				return err
			}
		}
	}

	if !isRecord {
		return ErrInvalidSource
	}

	return nil
}

func decodeBins(recordValue reflect.Value, v any) error {
	if recordValue.Kind() != reflect.Map {
		return nil // continue
	}

	targetValue, err := structValue(v)
	if err != nil {
		return err
	}

	targetType := targetValue.Type()
	for i := 0; i < targetType.NumField(); i++ {
		fieldValue := fieldValueDeref(targetValue, i)
		if fieldValue.Kind() == reflect.Struct {
			if err := decodeBins(recordValue, fieldValue); err != nil {
				return err
			}
			continue
		}

		aeroTag := targetType.Field(i).Tag.Get(mapperTag)
		if aeroTag == "" {
			continue
		}

		// parse the field tag
		tag, err := parseTag(aeroTag)
		if err != nil {
			return err
		}

		if tag.name == "" {
			continue
		}

		binValue := recordValue.MapIndex(reflect.ValueOf(tag.name))
		if binValue == reflectZeroValue { // not found
			continue
		}

		// check if the field can be set
		if !fieldValue.CanSet() {
			continue
		}

		// convert the source value to the correct type
		convertedValue, err := convertElementType(binValue, targetType.Field(i).Type)
		if err != nil {
			return fmt.Errorf("error converting value for field %s: %w",
				targetType.Field(i).Name, err)
		}

		// set the value
		fieldValue.Set(convertedValue)
	}

	return nil
}

//nolint:gocyclo,funlen
func decodeKey(recordValue reflect.Value, v any) error {
	targetValue, err := structValue(v)
	if err != nil {
		return err
	}

	targetType := targetValue.Type()
	for i := 0; i < targetType.NumField(); i++ {
		fieldValue := fieldValueDeref(targetValue, i)
		if fieldValue.Kind() == reflect.Struct {
			if err := decodeKey(recordValue, fieldValue); err != nil {
				return err
			}
			continue
		}

		aeroTag := targetType.Field(i).Tag.Get(mapperTag)
		if aeroTag == "" {
			continue
		}

		// parse the field tag
		tag, err := parseTag(aeroTag)
		if err != nil {
			return err
		}

		if !tag.meta {
			continue
		}

		switch tag.name {
		case metaTagNamespace:
			if !fieldValue.CanSet() {
				return fmt.Errorf("%s value cannot be changed", fieldValue.Type().Name())
			}

			m, err := getMethod(recordValue, "Namespace")
			if err != nil {
				return fmt.Errorf("%s: %w", recordValue.Type().Name(), err)
			}

			// call the method and convert the result correctly
			results := m.Call(nil)
			if len(results) == 0 {
				return fmt.Errorf("method Namespace returned no values")
			}

			if results[0].Type() == fieldValue.Type() {
				fieldValue.Set(results[0]) // set the return value
			} else {
				return fmt.Errorf("method Namespace returned wrong type. Expected %s, got %s",
					fieldValue.Type(), results[0].Type())
			}

		case metaTagSetName:
			if !fieldValue.CanSet() {
				return fmt.Errorf("%s value cannot be changed", fieldValue.Type().Name())
			}

			m, err := getMethod(recordValue, "SetName")
			if err != nil {
				return err
			}

			// call the method and convert the result correctly
			results := m.Call(nil)
			if len(results) == 0 {
				return fmt.Errorf("method SetName returned no values")
			}

			if results[0].Type() == fieldValue.Type() {
				fieldValue.Set(results[0]) // set the return value
			} else {
				return fmt.Errorf("method SetName returned wrong type. Expected %s, got %s",
					fieldValue.Type(), results[0].Type())
			}

		case metaTagUserKey:
			if !fieldValue.CanSet() {
				return fmt.Errorf("%s value cannot be changed", fieldValue.Type().Name())
			}

			m, err := getMethod(recordValue, "Value")
			if err != nil {
				return err
			}

			// call the method and convert the result correctly
			results := m.Call(nil)
			if len(results) == 0 {
				return fmt.Errorf("method Value returned no values")
			}

			userKeyValue := results[0]

			// call GetObject on the returned value
			methodToCall := userKeyValue.MethodByName("GetObject")
			if !methodToCall.IsValid() {
				return fmt.Errorf("method GetObject not found on type %s", userKeyValue.Type())
			}

			results = methodToCall.Call(nil)
			if len(results) == 0 {
				return fmt.Errorf("method GetObject returned no values")
			}

			if results[0].Type() == fieldValue.Type() {
				fieldValue.Set(results[0]) // set the return value
			} else {
				return fmt.Errorf("method Value returned wrong type. Expected %s, got %s",
					fieldValue.Type(), results[0].Type())
			}

		case metaTagDigest:
			if !fieldValue.CanSet() {
				return fmt.Errorf("%s value cannot be changed", fieldValue.Type().Name())
			}

			m, err := getMethod(recordValue, "Digest")
			if err != nil {
				return err
			}

			// call the method and convert the result correctly
			results := m.Call(nil)
			if len(results) == 0 {
				return fmt.Errorf("method Digest returned no values")
			}

			digestValue := results[0]
			digestBytes := digestValue.Bytes()

			switch {
			case fieldValue.Type().Kind() == reflect.Array &&
				fieldValue.Type().Elem().Kind() == reflect.Uint8:
				// check if the array has the correct length
				if fieldValue.Type().Len() != len(digestBytes) {
					return fmt.Errorf("method Digest returned []byte of length %d, "+
						"expected array of length %d", len(digestBytes), fieldValue.Type().Len())
				}

				// copy the bytes into the array
				for i := 0; i < fieldValue.Type().Len(); i++ {
					fieldValue.Index(i).Set(reflect.ValueOf(digestBytes[i]))
				}
			case fieldValue.Type() == reflect.TypeOf([]byte{}): // the field is a []byte
				fieldValue.Set(reflect.ValueOf(digestBytes))
			default:
				return fmt.Errorf("method Digest returned wrong type. Expected []byte or [N]byte, "+
					"got %s", fieldValue.Type())
			}
		}
	}

	return nil
}

func decodeRecord(recordValue reflect.Value, v any) error {
	targetValue, err := structValue(v)
	if err != nil {
		return err
	}

	targetType := targetValue.Type()
	for i := 0; i < targetType.NumField(); i++ {
		fieldValue := fieldValueDeref(targetValue, i)
		if fieldValue.Kind() == reflect.Struct {
			if err := decodeRecord(recordValue, fieldValue); err != nil {
				return err
			}
			continue
		}

		aeroTag := targetType.Field(i).Tag.Get(mapperTag)
		if aeroTag == "" {
			continue
		}

		// parse the field tag
		tag, err := parseTag(aeroTag)
		if err != nil {
			return err
		}

		if !tag.meta {
			continue
		}

		switch tag.name {
		case "generation":
			if !fieldValue.CanSet() {
				return fmt.Errorf("cannot set %s", tag.name)
			}

			f, err := getField(recordValue, "Generation")
			if err != nil {
				return err
			}

			if err := setIntegerValue(fieldValue, f); err != nil {
				return fmt.Errorf("%s: %w", tag.name, err)
			}
		case "expiration":
			if !fieldValue.CanSet() {
				return fmt.Errorf("cannot set %s", tag.name)
			}

			f, err := getField(recordValue, "Expiration")
			if err != nil {
				return err
			}

			if err := setIntegerValue(fieldValue, f); err != nil {
				return fmt.Errorf("%s: %w", tag.name, err)
			}
		}
	}

	return nil
}

// setIntegerValue sets numeric field value.
func setIntegerValue(fieldValue, recordValue reflect.Value) error {
	// determine the field Kind and convert accordingly to prevent overflow
	switch fieldValue.Kind() {
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		// check if the source value is in the range of the destination type before
		// conversion
		if fieldValue.OverflowInt(recordValue.Int()) {
			return fmt.Errorf("value '%d' overflows destination type '%s'",
				recordValue.Int(), fieldValue.Type())
		}

		fieldValue.Set(reflect.ValueOf(recordValue.Convert(fieldValue.Type()).Interface()))
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		// check if the source value is in the range of the destination type before
		// conversion
		if fieldValue.OverflowUint(recordValue.Uint()) {
			return fmt.Errorf("value '%d' overflows destination type '%s'",
				recordValue.Int(), fieldValue.Type())
		}

		fieldValue.Set(reflect.ValueOf(recordValue.Convert(fieldValue.Type()).Interface()))
	default:
		return fmt.Errorf("unsupported field type: %s", fieldValue.Type())
	}

	return nil
}

// parseTag parses the aero tag.
func parseTag(tagString string) (tag, error) {
	var parsed tag
	parts := strings.Split(tagString, ",")

	for _, p := range parts {
		part := strings.TrimSpace(p)
		switch part {
		case tagValueMeta:
			parsed.meta = true
		case tagValueOmitempty:
			parsed.omitempty = true
		case tagValueOmit:
			parsed.omit = true
		default:
			if parsed.name == "" {
				parsed.name = part
			} else {
				return tag{}, fmt.Errorf("invalid tag: %s", tagString)
			}
		}
	}

	// handle 'meta' bin name
	if parsed.meta && parsed.name == "" {
		parsed.meta = false
		parsed.name = tagValueMeta
	}

	return parsed, nil
}
