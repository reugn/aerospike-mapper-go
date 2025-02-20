package testtypes

import (
	"fmt"
	"reflect"
	"strconv"
)

// MapPair is used when the client returns sorted maps from the server
// Since the default map in Go is a hash map, we will use a slice
// to return the results in server order
type MapPair struct{ Key, Value interface{} }

// Value interface is used to efficiently serialize objects into the wire protocol.
type Value interface {
	// GetObject returns original value as an interface{}.
	GetObject() interface{}

	// String implements Stringer interface.
	String() string
}

// AerospikeBlob interface allows the user to write a conversion function from their value to []bytes.
type AerospikeBlob interface {
	// EncodeBlob returns a byte slice representing the encoding of the
	// receiver for transmission to a Decoder, usually of the same
	// concrete type.
	EncodeBlob() ([]byte, error)
}

// OpResults encapsulates the results of batch read operations
type OpResults []interface{}

// NewValue generates a new Value object based on the type.
// If the type is not supported, NewValue will panic.
func NewValue(v interface{}) Value {
	if res := concreteNewValueReflect(v); res != nil {
		return res
	}

	// panic for anything that is not supported.
	panic(fmt.Sprintf("Value type '%v' (%s) not supported",
		v, reflect.TypeOf(v).String()))
}

// NullValue is an empty value.
type NullValue struct{}

var nullValue NullValue

// NewNullValue generates a NullValue instance.
func NewNullValue() NullValue {
	return nullValue
}

// GetObject returns original value as an interface{}.
func (vl NullValue) GetObject() interface{} {
	return nil
}

func (vl NullValue) String() string {
	return ""
}

///////////////////////////////////////////////////////////////////////////////

// InfinityValue is an empty value.
type InfinityValue struct{}

var infinityValue InfinityValue

// NewInfinityValue generates a InfinityValue instance.
func NewInfinityValue() InfinityValue {
	return infinityValue
}

// GetObject returns original value as an interface{}.
func (vl InfinityValue) GetObject() interface{} {
	return nil
}

func (vl InfinityValue) String() string {
	return "INF"
}

///////////////////////////////////////////////////////////////////////////////

// WildCardValue is an empty value.
type WildCardValue struct{}

var wildCardValue WildCardValue

// NewWildCardValue generates a WildCardValue instance.
func NewWildCardValue() WildCardValue {
	return wildCardValue
}

// GetObject returns original value as an interface{}.
func (vl WildCardValue) GetObject() interface{} {
	return nil
}

func (vl WildCardValue) String() string {
	return "*"
}

///////////////////////////////////////////////////////////////////////////////

// BytesValue encapsulates an array of bytes.
type BytesValue []byte

// NewBytesValue generates a ByteValue instance.
func NewBytesValue(bytes []byte) BytesValue {
	return BytesValue(bytes)
}

// NewBlobValue accepts an AerospikeBlob interface, and automatically
// converts it to a BytesValue.
// If Encode returns an err, it will panic.
func NewBlobValue(object AerospikeBlob) BytesValue {
	buf, err := object.EncodeBlob()
	if err != nil {
		panic(err)
	}

	return NewBytesValue(buf)
}

// GetObject returns original value as an interface{}.
func (vl BytesValue) GetObject() interface{} {
	return []byte(vl)
}

// String implements Stringer interface.
func (vl BytesValue) String() string {
	return fmt.Sprintf("% 02x", []byte(vl))
}

///////////////////////////////////////////////////////////////////////////////

// StringValue encapsulates a string value.
type StringValue string

// NewStringValue generates a StringValue instance.
func NewStringValue(value string) StringValue {
	return StringValue(value)
}

// GetObject returns original value as an interface{}.
func (vl StringValue) GetObject() interface{} {
	return string(vl)
}

// String implements Stringer interface.
func (vl StringValue) String() string {
	return string(vl)
}

///////////////////////////////////////////////////////////////////////////////

// IntegerValue encapsulates an integer value.
type IntegerValue int

// NewIntegerValue generates an IntegerValue instance.
func NewIntegerValue(value int) IntegerValue {
	return IntegerValue(value)
}

// GetObject returns original value as an interface{}.
func (vl IntegerValue) GetObject() interface{} {
	return int(vl)
}

// String implements Stringer interface.
func (vl IntegerValue) String() string {
	return strconv.Itoa(int(vl))
}

///////////////////////////////////////////////////////////////////////////////

// LongValue encapsulates an int64 value.
type LongValue int64

// NewLongValue generates a LongValue instance.
func NewLongValue(value int64) LongValue {
	return LongValue(value)
}

// GetObject returns original value as an interface{}.
func (vl LongValue) GetObject() interface{} {
	return int64(vl)
}

// String implements Stringer interface.
func (vl LongValue) String() string {
	return strconv.Itoa(int(vl))
}

///////////////////////////////////////////////////////////////////////////////

// FloatValue encapsulates an float64 value.
type FloatValue float64

// NewFloatValue generates a FloatValue instance.
func NewFloatValue(value float64) FloatValue {
	return FloatValue(value)
}

// GetObject returns original value as an interface{}.
func (vl FloatValue) GetObject() interface{} {
	return float64(vl)
}

// String implements Stringer interface.
func (vl FloatValue) String() string {
	return (fmt.Sprintf("%f", vl))
}

///////////////////////////////////////////////////////////////////////////////

// BoolValue encapsulates a boolean value.
// Supported by Aerospike server v5.6+ only.
type BoolValue bool

// GetObject returns original value as an interface{}.
func (vb BoolValue) GetObject() interface{} {
	return bool(vb)
}

// String implements Stringer interface.
func (vb BoolValue) String() string {
	return (fmt.Sprintf("%v", bool(vb)))
}

///////////////////////////////////////////////////////////////////////////////

// ValueArray encapsulates an array of Value.
// Supported by Aerospike 3+ servers only.
type ValueArray []Value

// NewValueArray generates a ValueArray instance.
func NewValueArray(array []Value) *ValueArray {
	// return &ValueArray{*NewListerValue(valueList(array))}
	res := ValueArray(array)
	return &res
}

// GetObject returns original value as an interface{}.
func (va ValueArray) GetObject() interface{} {
	return []Value(va)
}

// String implements Stringer interface.
func (va ValueArray) String() string {
	return fmt.Sprintf("%v", []Value(va))
}

///////////////////////////////////////////////////////////////////////////////

// ListValue encapsulates any arbitrary array.
// Supported by Aerospike 3+ servers only.
type ListValue []interface{}

// NewListValue generates a ListValue instance.
func NewListValue(list []interface{}) ListValue {
	return ListValue(list)
}

// GetObject returns original value as an interface{}.
func (vl ListValue) GetObject() interface{} {
	return []interface{}(vl)
}

// String implements Stringer interface.
func (vl ListValue) String() string {
	return fmt.Sprintf("%v", []interface{}(vl))
}

///////////////////////////////////////////////////////////////////////////////

// MapValue encapsulates an arbitrary map.
// Supported by Aerospike 3+ servers only.
type MapValue map[interface{}]interface{}

// NewMapValue generates a MapValue instance.
func NewMapValue(vmap map[interface{}]interface{}) MapValue {
	return MapValue(vmap)
}

// GetObject returns original value as an interface{}.
func (vl MapValue) GetObject() interface{} {
	return map[interface{}]interface{}(vl)
}

func (vl MapValue) String() string {
	return fmt.Sprintf("%v", map[interface{}]interface{}(vl))
}

///////////////////////////////////////////////////////////////////////////////

// JsonValue encapsulates a Json map.
// Supported by Aerospike 3+ servers only.
type JsonValue map[string]interface{}

// NewJsonValue generates a JsonValue instance.
func NewJsonValue(vmap map[string]interface{}) JsonValue {
	return JsonValue(vmap)
}

// GetObject returns original value as an interface{}.
func (vl JsonValue) GetObject() interface{} {
	return map[string]interface{}(vl)
}

func (vl JsonValue) String() string {
	return fmt.Sprintf("%v", map[string]interface{}(vl))
}

///////////////////////////////////////////////////////////////////////////////

// GeoJSONValue encapsulates a 2D Geo point.
// Supported by Aerospike 3.6.1 servers and later only.
type GeoJSONValue string

// NewGeoJSONValue generates a GeoJSONValue instance.
func NewGeoJSONValue(value string) GeoJSONValue {
	res := GeoJSONValue(value)
	return res
}

// GetObject returns original value as an interface{}.
func (vl GeoJSONValue) GetObject() interface{} {
	return string(vl)
}

// String implements Stringer interface.
func (vl GeoJSONValue) String() string {
	return string(vl)
}

///////////////////////////////////////////////////////////////////////////////

// HLLValue encapsulates a HyperLogLog value.
type HLLValue []byte

// NewHLLValue generates a ByteValue instance.
func NewHLLValue(bytes []byte) HLLValue {
	return HLLValue(bytes)
}

// GetObject returns original value as an interface{}.
func (vl HLLValue) GetObject() interface{} {
	return []byte(vl)
}

// String implements Stringer interface.
func (vl HLLValue) String() string {
	return fmt.Sprintf("% 02x", []byte(vl))
}

///////////////////////////////////////////////////////////////////////////////

// RawBlobValue encapsulates a CDT BLOB value.
// Notice: Do not use this value, it is for internal aerospike use only.
type RawBlobValue struct {
	// ParticleType signifies the data
	ParticleType int
	// Data carries the data
	Data []byte
}

// NewRawBlobValue generates a RawBlobValue instance for a CDT List or map using a particle type.
func NewRawBlobValue(pt int, b []byte) *RawBlobValue {
	data := make([]byte, len(b))
	copy(data, b)
	return &RawBlobValue{ParticleType: pt, Data: data}
}

// GetObject returns original value as an interface{}.
func (vl *RawBlobValue) GetObject() interface{} {
	return vl.Data
}

// String implements Stringer interface.
func (vl *RawBlobValue) String() string {
	return fmt.Sprintf("% 02x", vl.Data)
}

///////////////////////////////////////////////////////////////////////////////

// if the returned value is nil, the caller will panic
//
//nolint:gosec
func concreteNewValueReflect(v interface{}) Value {
	// check for array and map
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
		l := rv.Len()
		arr := make([]interface{}, l)
		for i := 0; i < l; i++ {
			arr[i] = rv.Index(i).Interface()
		}

		return NewListValue(arr)
	case reflect.Map:
		l := rv.Len()
		amap := make(map[interface{}]interface{}, l)
		for _, i := range rv.MapKeys() {
			amap[i.Interface()] = rv.MapIndex(i).Interface()
		}

		return NewMapValue(amap)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return NewLongValue(reflect.ValueOf(v).Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return NewLongValue(int64(reflect.ValueOf(v).Uint()))
	case reflect.String:
		return NewStringValue(rv.String())
	}

	return nil
}
