package influxqu

import (
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

func getFiledAsString(val reflect.Value, i int) (string, error) {
	f := val.Field(i)

	if f.Kind() == reflect.Ptr {
		f = f.Elem()
	}

	if f.Kind() != reflect.String {
		return "", &UnSupportedType{}
	}

	return f.String(), nil
}

func getFiledAsTime(val reflect.Value, i int) (time.Time, error) {
	f := val.Field(i)

	if f.Kind() == reflect.Ptr {
		f = f.Elem()
	}

	t, ok := f.Interface().(time.Time)
	if !ok {
		return time.Time{}, &UnSupportedType{}
	}

	return t, nil
}
