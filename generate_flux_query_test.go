package influxqu

import (
	"testing"
	"time"
)

func Test_GenerateFluxQuery(t *testing.T) {
	type Data struct {
		Base      string    `influxqu:"measurement"`
		T1        string    `influxqu:"tag,t1"`
		T2        string    `influxqu:"tag,t2,omitempty"`
		F1        int       `influxqu:"field,f1"`
		F2        bool      `influxqu:"field,f2"`
		F3        string    `influxqu:"field,f3"`
		Timestamp time.Time `influxqu:"timestamp"`
	}

	g := NewinfluxQu()
	data := Data{
		Base:      "base",
		T1:        "abc",
		T2:        "",
		F1:        1,
		F2:        true,
		F3:        "",
		Timestamp: time.Now(),
	}

	exptected := `from(bucket: "bucket")
 |> range(start: -1h)
 |> filter(fn: (r) => r["t1"] == "abc")
 |> filter(fn: (r) => r["_measurement"] == "base")
 |> filter(fn: (r) => r["_field"] == "f1" or r["_field"] == "f2")
 |> sort("_time")
 |> last()`

	q, err := g.GenerateFluxQuery("bucket", "-1h", "", data, []string{"sort(\"_time\")", "last()"})
	if err != nil {
		t.Error(err)
	}

	if q != exptected {
		t.Error("query is not expected")
	}
}
