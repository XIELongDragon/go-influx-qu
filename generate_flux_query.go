package influxqu

import "reflect"

func (q *influxQu) generateFluxQuery(
	bucket, start, end string,
	tags map[string]string,
	fields []string,
	suffixes []string,
) (query string, cols []string) {
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
		cols = append(cols, k)
	}

	m := ""

	for i, f := range fields {
		if i != 0 {
			m += " or "
		}

		m += "r[\"_field\"] == \"" + f + "\""
		cols = append(cols, f)
	}

	if m != "" {
		query = query + "\n |> filter(fn: (r) => " + m + ")"
	}

	for _, s := range suffixes {
		query += "\n |> " + s
	}

	return query, cols
}

func (q *influxQu) GenerateFluxQuery(
	bucket, start, end string, v interface{}, suffixes []string,
) (query string, cols []string, err error) {
	val := reflect.Indirect(reflect.ValueOf(v))
	valType, valKind := getTypeInfo(v, val)

	if valKind != reflect.Struct {
		return "", nil, &UnSupportedType{}
	}

	measurement, tags, omitTags, f, _, err := q.getData(v, valType)
	if err != nil {
		return "", nil, err
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

	query, cols = q.generateFluxQuery(bucket, start, end, tags, fields, suffixes)
	cols = append(cols, omitTags...)

	return query, cols, nil
}
