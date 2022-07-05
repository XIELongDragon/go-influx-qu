package influxqu

import (
	"testing"
	"time"
)

func Test_influxQu_GenerateInfluxPoint(t *testing.T) {
	type Data struct {
		Base      string    `influxqu:"measurement"`
		T1        string    `influxqu:"tag,t1"`
		T2        string    `influxqu:"tag,t2"`
		F1        int       `influxqu:"field,f1"`
		F2        bool      `influxqu:"field,f2"`
		Timestamp time.Time `influxqu:"timestamp"`
	}

	g := NewinfluxQu()
	data := Data{
		Base:      "base",
		T1:        "t1",
		T2:        "t2",
		F1:        1,
		F2:        true,
		Timestamp: time.Now(),
	}

	p, e := g.GenerateInfluxPoint(&data)
	if e != nil {
		t.Error(e)
	}

	if p == nil {
		t.Error("point is nil")
	}

	if p.Name() != data.Base {
		t.Error("point name is not base")
	}

	if p.Time() != data.Timestamp {
		t.Error("point timestamp is not data.Timestamp")
	}

	for _, v := range p.TagList() {
		if v.Key == "t1" && v.Value != data.T1 {
			t.Errorf("tag t1 value is not t1, but %s", v.Value)
		}

		if v.Key == "t2" && v.Value != data.T2 {
			t.Errorf("tag t2 value is not t2, but %s", v.Value)
		}

		if v.Key != "t1" && v.Key != "t2" {
			t.Errorf("tag %s is not t1 or t2", v.Key)
		}
	}

	for _, v := range p.FieldList() {
		if v.Key != "f1" && v.Key != "f2" {
			t.Errorf("field %s is not f1 or f2", v.Key)
		}

		if v.Key == "f1" {
			if vi, ok := v.Value.(int64); !ok || vi != int64(data.F1) {
				t.Errorf("field f1 value is not 1, but %v", v.Value)
			}
		}

		if v.Key == "f2" {
			if vi, ok := v.Value.(bool); !ok || vi != data.F2 {
				t.Errorf("field f2 value is not true, but %v", v.Value)
			}
		}
	}
}
