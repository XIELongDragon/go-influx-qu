package influxqu

import (
	"reflect"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type influxQu struct {
	key            string
	measurementKey string
	fieldKey       string
	tagKey         string
	timestampKey   string
}

func (q *influxQu) getData(v interface{}, t reflect.Type) (measurement string, tags map[string]string, fields map[string]interface{}, timestamp *time.Time, err error) {
	val := reflect.Indirect(reflect.ValueOf(v))
	tags = make(map[string]string)
	fields = make(map[string]interface{})

	n := t.NumField()
	for i := 0; i < n; i++ {
		f := t.Field(i)
		if f.Anonymous && (f.Type.Kind() == reflect.Struct || f.Type.Kind() == reflect.Ptr) {
			m, t, f, tp, err := q.getData(val.Field(i).Interface(), f.Type)
			if err != nil {
				return "", nil, nil, nil, err
			}

			if m != "" {
				if measurement != "" {
					return "", nil, nil, nil, &DuplicatedMeasurement{}
				}

				measurement = m
			}

			if tp != nil {
				if timestamp != nil {
					return "", nil, nil, nil, &DuplicatedTimestamp{}
				}

				timestamp = tp
			}

			for k, v := range t {
				if _, ok := tags[k]; ok {
					return "", nil, nil, nil, &DuplicatedTag{}
				}

				tags[k] = v
			}

			for k, v := range f {
				if _, ok := fields[k]; ok {
					return "", nil, nil, nil, &DuplicatedField{}
				}

				fields[k] = v
			}
		}

		tag := f.Tag.Get(q.key)
		if tag == "" {
			continue
		}

		tgs := strings.Split(tag, ",")
		for i := range tgs {
			tgs[i] = strings.TrimSpace(tgs[i])
		}

		switch tgs[0] {
		case q.measurementKey:
			if measurement != "" {
				return "", nil, nil, nil, &DuplicatedMeasurement{}
			}

			measurement, err = getFiledAsString(val, i)
			if err != nil {
				return "", nil, nil, nil, err
			}
		case q.tagKey:
			if len(tgs) < 2 {
				return "", nil, nil, nil, &NoTagName{}
			}

			t := tgs[1]

			if _, ok := tags[t]; ok {
				return "", nil, nil, nil, &DuplicatedTag{}
			}

			tags[t], err = getFiledAsString(val, i)
			if err != nil {
				return "", nil, nil, nil, err
			}
		case q.fieldKey:
			if len(tgs) < 2 {
				return "", nil, nil, nil, &NoFieldName{}
			}

			f := tgs[1]

			if _, ok := fields[f]; ok {
				return "", nil, nil, nil, &DuplicatedField{}
			}

			fields[f] = val.Field(i).Interface()

		case q.timestampKey:
			if timestamp != nil {
				return "", nil, nil, nil, &DuplicatedTimestamp{}
			}

			var tmp time.Time
			tmp, err = getFiledAsTime(val, i)
			if err != nil {
				return "", nil, nil, nil, err
			}

			timestamp = &tmp
		}

	}

	return measurement, tags, fields, timestamp, nil
}

func (q *influxQu) GenerateInfluxPoint(v interface{}) (*write.Point, error) {
	val := reflect.Indirect(reflect.ValueOf(v))
	valType, valKind := getTypeInfo(v, val)
	if valKind != reflect.Struct {
		return nil, &UnSupportedType{}
	}

	m, t, f, tp, err := q.getData(v, valType)
	if err != nil {
		return nil, err
	}

	if tp == nil {
		tm := time.Now()
		tp = &tm
	}

	return influxdb2.NewPoint(
		m, t, f, *tp,
	), nil
}
