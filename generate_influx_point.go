package influxqu

import (
	"reflect"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

func mergeTags(org, src map[string]string) error {
	for k, v := range src {
		if _, ok := org[k]; ok {
			return &DuplicatedTag{}
		}

		org[k] = v
	}

	return nil
}

func mergeFields(org, src map[string]interface{}) error {
	for k, v := range src {
		if _, ok := org[k]; ok {
			return &DuplicatedField{}
		}

		org[k] = v
	}

	return nil
}

func processTag(tags []string, org map[string]string, val reflect.Value, i int) (omiteTag string, err error) {
	if len(tags) < 2 {
		return "", &NoTagName{}
	}

	isOmitempty := false

	if len(tags) == 3 {
		if tags[2] != omitemptyKey {
			return "", &UnSupportedTag{}
		}

		isOmitempty = true
	}

	t := tags[1]

	if _, ok := org[t]; ok {
		return "", &DuplicatedTag{}
	}

	if val.IsValid() && val.Field(i).IsZero() {
		if isOmitempty {
			return t, nil
		}
	}

	v, err := getFieldAsString(val, i)
	if err != nil {
		return "", err
	}

	if !isOmitempty || v != "" {
		org[t] = v
	}

	return "", nil
}

func processFields(tags []string, org map[string]interface{}, val reflect.Value, i int) (err error) {
	if len(tags) < 2 {
		return &NoFieldName{}
	}

	isOmitempty := false

	if len(tags) == 3 {
		if tags[2] != omitemptyKey {
			return &UnSupportedTag{}
		}

		isOmitempty = true
	}

	f := tags[1]

	if _, ok := org[f]; ok {
		return &DuplicatedField{}
	}

	v := val.Field(i).Interface()
	if !isOmitempty || !isValueEmpty(v) {
		org[f] = v
	}

	return nil
}

func (q *influxQu) processSubStruct(
	val interface{},
	ty reflect.Type,
	measurement *string,
	tags map[string]string,
	fields map[string]interface{},
	timestamp *time.Time) error {
	m, t, _, f, tp, e := q.getData(val, ty)
	if e != nil {
		return e
	}

	if m != "" {
		if *measurement != "" {
			return &DuplicatedMeasurement{}
		}

		*measurement = m
	}

	if tp != nil {
		if timestamp != nil {
			return &DuplicatedTimestamp{}
		}

		*timestamp = *tp
	}

	if err := mergeTags(tags, t); err != nil {
		return err
	}

	if err := mergeFields(fields, f); err != nil {
		return err
	}

	return nil
}

func (q *influxQu) getData(v interface{}, t reflect.Type) (
	measurement string,
	tags map[string]string,
	omiteTags []string,
	fields map[string]interface{},
	timestamp *time.Time,
	err error,
) {
	val := reflect.Indirect(reflect.ValueOf(v))
	tags = make(map[string]string)
	fields = make(map[string]interface{})
	omiteTags = make([]string, 0)

	n := t.NumField()
	for i := 0; i < n; i++ {
		f := t.Field(i)
		if f.Anonymous && (f.Type.Kind() == reflect.Struct || f.Type.Kind() == reflect.Ptr) {
			err = q.processSubStruct(val.Field(i).Interface(), f.Type, &measurement, tags, fields, timestamp)
			if err != nil {
				return "", nil, nil, nil, nil, err
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
				return "", nil, nil, nil, nil, &DuplicatedMeasurement{}
			}

			if len(tgs) != 1 {
				return "", nil, nil, nil, nil, &UnSupportedTag{}
			}

			measurement, err = getFieldAsString(val, i)
			if err != nil {
				return "", nil, nil, nil, nil, err
			}
		case q.tagKey:
			omiteTag, err := processTag(tgs, tags, val, i)
			if err != nil {
				return "", nil, nil, nil, nil, err
			}
			if omiteTag != "" {
				omiteTags = append(omiteTags, omiteTag)
			}
		case q.fieldKey:
			err = processFields(tgs, fields, val, i)
			if err != nil {
				return "", nil, nil, nil, nil, err
			}

		case q.timestampKey:
			if timestamp != nil {
				return "", nil, nil, nil, nil, &DuplicatedTimestamp{}
			}

			var tmp time.Time
			tmp, err = getFiledAsTime(val, i)

			if err != nil {
				return "", nil, nil, nil, nil, err
			}

			timestamp = &tmp
		}
	}

	return measurement, tags, omiteTags, fields, timestamp, nil
}

func (q *influxQu) GenerateInfluxPoint(v interface{}) (*write.Point, error) {
	val := reflect.Indirect(reflect.ValueOf(v))
	valType, valKind := getTypeInfo(v, val)

	if valKind != reflect.Struct {
		return nil, &UnSupportedType{}
	}

	m, t, _, f, tp, err := q.getData(v, valType)
	if err != nil {
		return nil, err
	}

	if m == "" {
		return nil, &NoValidMeasurement{}
	}

	if len(f) == 0 {
		return nil, &NoValidField{}
	}

	if tp == nil {
		tm := time.Now()
		tp = &tm
	}

	return influxdb2.NewPoint(
		m, t, f, *tp,
	), nil
}
