package assert

import (
	"errors"
	"reflect"
	"testing"
)

// Equal verifies equality of two objects.
func Equal[T any](t *testing.T, a T, b T) {
	if !reflect.DeepEqual(a, b) {
		t.Helper()
		t.Fatalf("%v != %v", a, b)
	}
}

// IsNil verifies that the object is nil.
func IsNil(t *testing.T, obj any) {
	if obj != nil {
		value := reflect.ValueOf(obj)
		switch value.Kind() {
		case reflect.Ptr, reflect.Map, reflect.Slice,
			reflect.Interface, reflect.Func, reflect.Chan:
			if value.IsNil() {
				return
			}
		default:
		}
		t.Helper()
		t.Fatalf("%v is not nil", obj)
	}
}

// ErrorIs checks whether any error in err's tree matches target.
func ErrorIs(t *testing.T, err error, target error) {
	if !errors.Is(err, target) {
		t.Helper()
		t.Fatalf("Error type mismatch: %v != %v", err, target)
	}
}
