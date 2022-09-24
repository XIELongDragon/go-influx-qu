package influxqu

import "reflect"

func (q *influxQu) generateFluxQuery(
	bucket, start, end string,
	tags map[string]string,
	fields []string,
	suffixes []string) string {
	var query string
	query = "from(bucket: \"" + bucket + "\")"

	if start != "" && end != "" {
		query += "\n |> range(start: " + start + ", stop: " + end + ")"
	} else if start != "" {
		query += "\n |> range(start: " + start + ")"
	} else if end != "" {
		query += "\n |> range(stop: " + end + ")"
	}

	for k, v := range tags {
		query = query + "\n |> filter(fn: (r) => r[\"" + k + "\"] == \"" + v + "\")"
	}

	m := ""

	for i, f := range fields {
		if i != 0 {
			m += " or "
		}

		m += "r[\"_field\"] == \"" + f + "\""
	}

	if m != "" {
		query = query + "\n |> filter(fn: (r) => " + m + ")"
	}

	for _, s := range suffixes {
		query += "\n |> " + s
	}

	return query
}

func (q *influxQu) GenerateFluxQuery(bucket, start, end string, v interface{}, suffixes []string) (string, error) {
	val := reflect.Indirect(reflect.ValueOf(v))
	valType, valKind := getTypeInfo(v, val)

	if valKind != reflect.Struct {
		return "", &UnSupportedType{}
	}

	measurement, tags, f, _, err := q.getData(v, valType)
	if err != nil {
		return "", err
	}

	if measurement != "" {
		tags["_measurement"] = measurement
	}

	fields := make([]string, 0, len(f))
	for k, v := range f {
		if !isValueEmpty(v) {
			fields = append(fields, k)
		}
	}

	return q.generateFluxQuery(bucket, start, end, tags, fields, suffixes), nil
}
