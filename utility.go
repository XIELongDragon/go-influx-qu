package influxqu

import (
	"encoding"
	"fmt"
	"reflect"
	"time"
)

func isValueEmpty(val interface{}) bool {
	if val == nil {
		return true
	}

	if reflect.TypeOf(val).Kind() == reflect.Ptr {
		return reflect.ValueOf(val).IsNil()
	}

	v := reflect.ValueOf(val)

	return v.IsValid() && v.IsZero()
}

func getTypeInfo(i interface{}, val reflect.Value) (reflect.Type, reflect.Kind) {
	var t reflect.Type
	valKind := val.Kind()

	if valKind == reflect.Slice {
		if reflect.ValueOf(i).Kind() == reflect.Ptr {
			t = reflect.TypeOf(i).Elem().Elem()
		} else {
			t = reflect.TypeOf(i).Elem()
		}

		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		valKind = t.Kind()
	} else {
		t = val.Type()
	}

	return t, valKind
}

func getFieldAsString(val reflect.Value, i int) (string, error) {
	f := val.Field(i)

	if f.Kind() == reflect.Ptr {
		if f.IsNil() {
			return "", nil
		}

		f = f.Elem()
	}

	if f.Kind() == reflect.String {
		return f.String(), nil
	}

	if s, ok := f.Interface().(fmt.Stringer); ok {
		return s.String(), nil
	}

	if s, ok := f.Interface().(encoding.TextMarshaler); ok {
		b, err := s.MarshalText()
		if err != nil {
			return "", err
		}

		return string(b), nil
	}

	switch v := f.Interface().(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v), nil
	case float32, float64:
		return fmt.Sprintf("%f", v), nil
	case bool:
		return fmt.Sprintf("%t", v), nil
	case time.Time:
		return v.Format(time.RFC3339Nano), nil
	}

	return "", &UnSupportedType{}
}

func getFiledAsTime(val reflect.Value, i int) (time.Time, error) {
	f := val.Field(i)

	if f.Kind() == reflect.Ptr {
		if f.IsNil() {
			return time.Time{}, &UnSupportedType{}
		}

		f = f.Elem()
	}

	t, ok := f.Interface().(time.Time)
	if !ok {
		return time.Time{}, &UnSupportedType{}
	}

	return t, nil
}
