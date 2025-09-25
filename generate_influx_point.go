package influxqu

import (
	"reflect"
	"strings"
	"time"

	"github.com/InfluxCommunity/influxdb3-go/v2/influxdb3"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/shopspring/decimal"
)

const (
	decimalPkgPath    = "github.com/shopspring/decimal"
	decimalStructName = "Decimal"
)

func mergeOmitTags(org, src []string) ([]string, error) {
	tmp := map[string]struct{}{}

	for _, t := range org {
		tmp[t] = struct{}{}
	}

	for _, t := range src {
		if _, ok := tmp[t]; ok {
			return nil, &DuplicatedTag{}
		}

		tmp[t] = struct{}{}

		org = append(org, t)
	}

	return org, nil
}

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
		// process decimal
		if val.Field(i).Type().PkgPath() == decimalPkgPath && val.Field(i).Type().Name() == decimalStructName {
			org[f] = v.(decimal.Decimal).InexactFloat64()
		} else if val.Field(i).Kind() == reflect.Pointer {
			underlyingVal := reflect.Indirect(val.Field(i))
			org[f] = underlyingVal.Interface()
		} else {
			org[f] = v
		}
	}

	return nil
}

func (q *influxQu) processSubStruct(
	val interface{},
	ty reflect.Type,
	measurement *string,
	tags map[string]string,
	omiteTags *[]string,
	fields map[string]interface{},
	timestamp *time.Time) (*time.Time, error) {
	m, t, o, f, tp, e := q.getData(val, ty)
	if e != nil {
		return nil, e
	}

	if m != "" {
		if *measurement != "" {
			return nil, &DuplicatedMeasurement{}
		}

		*measurement = m
	}

	if tp != nil {
		if timestamp != nil {
			return nil, &DuplicatedTimestamp{}
		}
	}

	if err := mergeTags(tags, t); err != nil {
		return nil, err
	}

	if *omiteTags, e = mergeOmitTags(*omiteTags, o); e != nil {
		return nil, e
	}

	if err := mergeFields(fields, f); err != nil {
		return nil, err
	}

	return tp, nil
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
			timestamp, err = q.processSubStruct(val.Field(i).Interface(), f.Type, &measurement, tags, &omiteTags, fields, timestamp)
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
			omiteTag, er := processTag(tgs, tags, val, i)
			if er != nil {
				return "", nil, nil, nil, nil, er
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
			tmp, err = getFieldAsTime(val, i)

			if err != nil {
				return "", nil, nil, nil, nil, err
			}

			timestamp = &tmp
		}
	}

	return measurement, tags, omiteTags, fields, timestamp, nil
}

func (q *influxQu) generateCommonPointInfo(v any) (
	measurement string,
	tags map[string]string,
	fields map[string]any,
	timestamp time.Time,
	err error,
) {
	val := reflect.Indirect(reflect.ValueOf(v))
	valType, valKind := getTypeInfo(v, val)

	if valKind != reflect.Struct {
		return "", nil, nil, time.Time{}, &UnSupportedType{}
	}

	m, t, _, f, tp, err := q.getData(v, valType)
	if err != nil {
		return "", nil, nil, time.Time{}, err
	}

	if m == "" {
		return "", nil, nil, time.Time{}, &NoValidMeasurement{}
	}

	if len(f) == 0 {
		return "", nil, nil, time.Time{}, &NoValidField{}
	}

	if tp == nil {
		tm := time.Now()
		tp = &tm
	}

	return m, t, f, *tp, nil
}

func (q *influxQu) GenerateInfluxPoint(v any) (*write.Point, error) {
	m, t, f, tp, err := q.generateCommonPointInfo(v)
	if err != nil {
		return nil, err
	}

	return influxdb2.NewPoint(m, t, f, tp), nil
}

func (q *influxQu) GenerateInfluxPointV3(v any) (*influxdb3.Point, error) {
	m, t, f, tp, err := q.generateCommonPointInfo(v)
	if err != nil {
		return nil, err
	}

	return influxdb3.NewPoint(m, t, f, tp), nil
}
