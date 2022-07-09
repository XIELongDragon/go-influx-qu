# go-influx-qu
A simple library converts structures to InfluxDB point and covert back from InfluxDB records to structures

## How to use it
1. Tag your structure fields, 
2. Call `GenerateInfluxPoint` to conver your structure to InfluxDB points

>**Note**
>
>if there i **NO timestamp tag** proived, will use current time as the point's timestamp

```go
type Data struct {
    Base      string    `influxqu:"measurement"`
	T1        string    `influxqu:"tag,t1"`
	T2        string    `influxqu:"tag,t2"`
	F1        int       `influxqu:"field,f1"`
	F2        bool      `influxqu:"field,f2"`
	Timestamp time.Time `influxqu:"timestamp"`

}

func main() {
    data := Data{
		Base:      "base",
		T1:        "t1",
		T2:        "t2",
		F1:        1,
		F2:        true,
		Timestamp: time.Now(),
	}

    g := NewinfluxQu()
    p, e := g.GenerateInfluxPoint(&data)

    // balbal
}
```
