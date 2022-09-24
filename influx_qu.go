package influxqu

import (
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type InfluxQu interface {
	GenerateInfluxPoint(val interface{}) (*write.Point, error)
	GenerateFluxQuery(bucket, start, end string, val interface{}, suffix []string) (string, error)
}

const (
	omitemptyKey = "omitempty"
)

type influxQu struct {
	key            string
	measurementKey string
	fieldKey       string
	tagKey         string
	timestampKey   string
}

func NewinfluxQu() InfluxQu {
	i, _ := NewinfluxQuWithKeys("influxqu", "measurement", "tag", "field", "timestamp")
	return i
}

func NewinfluxQuWithKeys(key string, measurementKey string, tagKey string, fieldKey string, timestampKey string) (InfluxQu, error) {
	if key == "" {
		key = "influxqu"
	}

	if measurementKey == "" {
		measurementKey = "measurement"
	}

	if fieldKey == "" {
		fieldKey = "field"
	}

	if tagKey == "" {
		tagKey = "tag"
	}

	if timestampKey == "" {
		timestampKey = "timestamp"
	}

	keys := make(map[string]struct{}, 5)

	keys[key] = struct{}{}
	keys[measurementKey] = struct{}{}
	keys[fieldKey] = struct{}{}
	keys[tagKey] = struct{}{}
	keys[timestampKey] = struct{}{}

	if len(keys) != 5 {
		return nil, &DuplicatedKey{}
	}

	return &influxQu{
		key:            key,
		measurementKey: measurementKey,
		fieldKey:       fieldKey,
		tagKey:         tagKey,
		timestampKey:   timestampKey,
	}, nil
}
