package influxqu

import (
	"testing"
	"time"
)

func Test_GenerateFluxQuery(t *testing.T) {
	type Data struct {
		Base string  `influxqu:"measurement"`
		T1   string  `influxqu:"tag,t1"`
		T2   string  `influxqu:"tag,t2,omitempty"`
		T3   bool    `influxqu:"tag,t3,omitempty"`
		T4   float32 `influxqu:"tag,t4,omitempty"`
		T5   float64 `influxqu:"tag,t5,omitempty"`
		T6   int     `influxqu:"tag,t6,omitempty"`
		F1   int     `influxqu:"field,f1"`
		F2   bool    `influxqu:"field,f2"`
		F3   string  `influxqu:"field,f3"`

		Timestamp time.Time `influxqu:"timestamp"`
	}

	g := NewinfluxQu()
	data := Data{
		Base:      "base",
		T1:        "abc",
		T2:        "",    // zero value and should be omitted
		T3:        false, // zero value and should be omitted
		T4:        0,     // zero value and should be omitted
		T5:        0,     // zero value and should be omitted
		T6:        0,     // zero value and should be omitted
		F1:        1,
		F2:        true,
		F3:        "",
		Timestamp: time.Now(),
	}

	expected := `from(bucket: "bucket")
 |> range(start: -1h)
 |> filter(fn: (r) => r["t1"] == "abc")
 |> filter(fn: (r) => r["_measurement"] == "base")
 |> filter(fn: (r) => r["_field"] == "f1" or r["_field"] == "f2")
 |> sort("_time")
 |> last()`

	q, cols, err := g.GenerateFluxQuery("bucket", "-1h", "", data, []string{"sort(\"_time\")", "last()"})
	if err != nil {
		t.Error(err)
	}

	if q != expected {
		t.Errorf("query is not expected, got: %s, expected: %s", q, expected)
	}

	expectedCols := []string{"t1", "_measurement", "f1", "f2", "t2", "t3", "t4", "t5", "t6"}

	if len(cols) != len(expectedCols) {
		t.Errorf("columns are not expected, got: %v, expected: %v", cols, expectedCols)
	}

	for i, col := range cols {
		if col != expectedCols[i] {
			t.Errorf("columns are not expected, got: %v, expected: %v", cols, expectedCols)
		}
	}
}
