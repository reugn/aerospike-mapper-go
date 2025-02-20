package mapper

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// fieldValueDeref returns field i of the given struct value.
// The returned value is dereferenced if it is a pointer.
func fieldValueDeref(value reflect.Value, i int) reflect.Value {
	fieldValue := value.Field(i)
	if fieldValue.Kind() == reflect.Ptr {
		fieldValue = fieldValue.Elem()
	}
	return fieldValue
}

// getMethod returns a field with the given name for the value.
func getField(value reflect.Value, fieldName string) (reflect.Value, error) {
	field := value.FieldByName(fieldName)
	if field.IsValid() {
		// field found directly on the value
		return field, nil
	}

	// if not found directly, try getting the field on a pointer to the value
	if value.CanAddr() { // check if the value is addressable
		ptrValue := value.Addr()
		// dereference the pointer to get the value
		field = ptrValue.Elem().FieldByName(fieldName)
		if field.IsValid() {
			// field found on the pointer to the value
			return field, nil
		}
	}

	// field not found
	return reflect.Value{}, fmt.Errorf("field '%s' not found on type %s (or pointer to it)",
		fieldName, value.Type())
}

// getMethod returns a method with the given name for the value.
func getMethod(value reflect.Value, methodName string) (reflect.Value, error) {
	method := value.MethodByName(methodName)
	if method.IsValid() {
		// method found directly on the value
		return method, nil
	}

	// if not found directly, try getting the method on a pointer to the value
	if value.CanAddr() { // check if the value is addressable
		ptrValue := value.Addr()
		method = ptrValue.MethodByName(methodName)
		if method.IsValid() {
			// method found on the pointer to the value
			return method, nil
		}
	}

	// method not found
	return reflect.Value{}, fmt.Errorf("method '%s' not found on type %s (or pointer to it)",
		methodName, value.Type())
}

// structValue returns a struct value for the given argument.
func structValue(value any) (reflect.Value, error) {
	var structValue reflect.Value
	switch v := value.(type) {
	case reflect.Value:
		structValue = v
	default:
		structValue = reflect.ValueOf(value)
		// check if value is a struct or pointer to a struct
		if structValue.Kind() == reflect.Ptr {
			// get the underlying struct value
			structValue = structValue.Elem()
		}

		// check if the underlying value is a struct
		if structValue.Kind() != reflect.Struct {
			return reflect.Value{}, ErrInvalidSourceType
		}
	}

	return structValue, nil
}

// isEmptyValue reports whether v is the zero value for its type.
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String, reflect.Chan,
		reflect.Func:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		if v.Type().String() == timeType {
			return v.IsZero()
		}
	case reflect.Invalid:
		return true
	default:
	}

	return false // consider as non-empty for other types
}

// convertElementType converts a value from a source field to the target type.
//
//nolint:gocyclo,funlen
func convertElementType(source any, targetType reflect.Type) (reflect.Value, error) {
	var sourceValue reflect.Value

	switch v := source.(type) {
	case reflect.Value:
		sourceValue = v
	default:
		sourceValue = reflect.ValueOf(source)
	}

	// if the source is of any type, get the underlying value
	if sourceValue.Kind() == reflect.Interface {
		sourceValue = sourceValue.Elem()
	}

	if !sourceValue.IsValid() {
		// handle invalid (zero) values, especially important for uninitialized fields
		// if the target is a pointer, return nil, otherwise return zero
		if targetType.Kind() == reflect.Ptr {
			return reflect.Zero(targetType), nil
		}
		return reflect.Zero(targetType), nil // return zero value for the target type
	}

	// dereference any pointers on the source
	if sourceValue.Kind() == reflect.Ptr {
		if sourceValue.IsNil() { // handle nil pointers
			if targetType.Kind() == reflect.Ptr {
				return reflect.Zero(targetType), nil // target is a pointer, set to nil
			}
			return reflect.Zero(targetType), nil // target is not a pointer, set to zero value
		}
		sourceValue = sourceValue.Elem() // dereference the pointer
	}

	sourceType := sourceValue.Type()

	// if the source and target types are already the same, return the source value
	if sourceType == targetType {
		return sourceValue, nil
	}

	switch targetType.Kind() {
	case reflect.String:
		// convert various types to string
		switch sourceType.Kind() {
		case reflect.String:
			return sourceValue, nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return reflect.ValueOf(strconv.FormatInt(sourceValue.Int(), 10)), nil
		case reflect.Float32, reflect.Float64:
			return reflect.ValueOf(strconv.FormatFloat(sourceValue.Float(), 'f', -1, 64)), nil
		case reflect.Bool:
			return reflect.ValueOf(strconv.FormatBool(sourceValue.Bool())), nil
		default:
			return reflect.Value{}, fmt.Errorf("cannot convert %s to string", sourceType.String())
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// convert various types to int
		switch sourceType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return sourceValue.Convert(targetType), nil // direct conversion is possible
		case reflect.Float32, reflect.Float64:
			return reflect.ValueOf(int64(sourceValue.Float())).Convert(targetType), nil
		case reflect.String:
			i, err := strconv.ParseInt(sourceValue.String(), 10, 64)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("cannot convert string '%s' to int: %w",
					sourceValue.String(), err)
			}
			return reflect.ValueOf(i).Convert(targetType), nil
		default:
			return reflect.Value{}, fmt.Errorf("cannot convert %s to int", sourceType.String())
		}

	case reflect.Float32, reflect.Float64:
		// convert various types to float
		switch sourceType.Kind() {
		case reflect.Float32, reflect.Float64:
			return sourceValue.Convert(targetType), nil // direct conversion
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return reflect.ValueOf(float64(sourceValue.Int())).Convert(targetType), nil
		case reflect.String:
			f, err := strconv.ParseFloat(sourceValue.String(), 64)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("cannot convert string '%s' to float: %w",
					sourceValue.String(), err)
			}
			return reflect.ValueOf(f).Convert(targetType), nil
		default:
			return reflect.Value{}, fmt.Errorf("cannot convert %s to float", sourceType.String())
		}

	case reflect.Bool:
		// convert various types to bool
		switch sourceType.Kind() {
		case reflect.Bool:
			return sourceValue, nil
		case reflect.String:
			b, err := strconv.ParseBool(sourceValue.String())
			if err != nil {
				return reflect.Value{}, fmt.Errorf("cannot convert string '%s' to bool: %w",
					sourceValue.String(), err)
			}
			return reflect.ValueOf(b), nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			// consider any non-zero integer as true
			return reflect.ValueOf(sourceValue.Int() != 0), nil
		case reflect.Float32, reflect.Float64:
			// consider any non-zero float as true
			return reflect.ValueOf(sourceValue.Float() != 0.0), nil
		default:
			return reflect.Value{}, fmt.Errorf("cannot convert %s to bool", sourceType.String())
		}

	case reflect.Slice:
		// handle slice conversion; requires element-by-element conversion
		if sourceType.Kind() != reflect.Slice {
			return reflect.Value{}, fmt.Errorf("cannot convert %s to slice", sourceType.String())
		}

		sourceLen := sourceValue.Len()
		elementType := targetType.Elem()
		newSlice := reflect.MakeSlice(targetType, sourceLen, sourceLen)

		for i := 0; i < sourceLen; i++ {
			sourceElement := sourceValue.Index(i)
			convertedElement, err := convertElementType(sourceElement.Interface(), elementType)
			if err != nil {
				return reflect.Value{},
					fmt.Errorf("error converting slice element at index %d: %w", i, err)
			}
			newSlice.Index(i).Set(convertedElement)
		}
		return newSlice, nil

	case reflect.Map:
		// handle map conversion; requires key and value conversion
		if sourceType.Kind() != reflect.Map {
			return reflect.Value{}, fmt.Errorf("cannot convert %s to map", sourceType.String())
		}

		keyType := targetType.Key()
		elementType := targetType.Elem()
		newMap := reflect.MakeMap(targetType)

		for _, key := range sourceValue.MapKeys() {
			sourceElement := sourceValue.MapIndex(key)

			convertedKey, err := convertElementType(key.Interface(), keyType)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("error converting map key: %w", err)
			}

			convertedValue, err := convertElementType(sourceElement.Interface(), elementType)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("error converting map value: %w", err)
			}

			newMap.SetMapIndex(convertedKey, convertedValue)
		}
		return newMap, nil

	case reflect.Struct:
		if targetType.String() == timeType {
			// attempt to convert to time.Time
			switch sourceType.Kind() {
			case reflect.String:
				// try parsing from string
				t, err := time.Parse(time.RFC3339, sourceValue.String()) // adjust layout as needed
				if err != nil {
					return reflect.Value{}, fmt.Errorf("cannot convert string '%s' to %s: %w",
						sourceValue.String(), timeType, err)
				}
				return reflect.ValueOf(t), nil

			case reflect.Struct:
				// if source is also a time.Time, perform a direct conversion
				if sourceType.String() == timeType {
					return sourceValue, nil
				}
				return reflect.Value{}, fmt.Errorf("cannot convert struct %s to %s",
					sourceType.String(), timeType)

			default:
				return reflect.Value{}, fmt.Errorf("cannot convert %s to %s",
					sourceType.String(), timeType)
			}
		} else { // handle nested structs; recursively convert the nested struct
			// create a new instance of the target struct
			nestedValue := reflect.New(targetType).Elem()

			// use a function to copy fields between structs
			err := copyStruct(source, nestedValue.Addr().Interface())
			if err != nil {
				return reflect.Value{}, fmt.Errorf("error mapping nested struct: %w", err)
			}
			return nestedValue, nil
		}

	case reflect.Ptr:
		// handle pointer conversion; create a new pointer to the target type
		// and recursively convert the underlying value
		targetElemType := targetType.Elem() // get the type the pointer points to

		convertedValue, err := convertElementType(source, targetElemType)
		if err != nil {
			return reflect.Value{}, err
		}

		// create a pointer to a new value of the converted type and set it
		newPtr := reflect.New(targetElemType)
		newPtr.Elem().Set(convertedValue)
		return newPtr, nil

	default:
		return reflect.Value{}, fmt.Errorf("unsupported target type: %s", targetType.Kind())
	}
}

// copyStruct copies values from one struct to another, handling different field names.
func copyStruct(source any, target any) error {
	sourceValue := reflect.ValueOf(source)

	// if the source is of any type, get the underlying value
	if sourceValue.Kind() == reflect.Interface {
		sourceValue = sourceValue.Elem()
	}

	if sourceValue.Kind() == reflect.Ptr {
		sourceValue = sourceValue.Elem()
	}

	targetValue := reflect.ValueOf(target)

	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer to a struct")
	}

	targetValue = targetValue.Elem() // dereference the pointer

	sourceType := sourceValue.Type()
	targetType := targetValue.Type()

	if sourceType.Kind() != reflect.Struct || targetType.Kind() != reflect.Struct {
		return fmt.Errorf("source and target must be structs")
	}

	for i := 0; i < sourceType.NumField(); i++ {
		sourceField := sourceType.Field(i)
		sourceFieldValue := sourceValue.Field(i)

		// use the aero tag to find the matching field in the target struct
		aeroTag := sourceField.Tag.Get(mapperTag)
		tag, err := parseTag(aeroTag)
		if err != nil {
			return err
		}

		if tag.name == "" {
			tag.name = sourceField.Name // fallback to the field name
		}

		// find the corresponding field in the target struct
		targetFieldValue := targetValue.FieldByName(tag.name)

		if !targetFieldValue.IsValid() || !targetFieldValue.CanSet() {
			continue // skip the field if not found or not settable
		}

		convertedValue, err := convertElementType(
			sourceFieldValue.Interface(),
			targetFieldValue.Type(),
		)
		if err != nil {
			return fmt.Errorf("error converting field %s: %w", tag.name, err)
		}

		targetFieldValue.Set(convertedValue)
	}

	return nil
}
