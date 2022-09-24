package influxqu

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

func checkTags(p *write.Point, targets map[string]string) error {
	for _, v := range p.TagList() {
		if tv, ok := targets[v.Key]; !ok {
			return fmt.Errorf("tag %s is not in target", v.Key)
		} else if tv != v.Value {
			return fmt.Errorf("tag %s value is not %s, but %s", v.Key, tv, v.Value)
		}
	}

	return nil
}

func checkFields(p *write.Point, targets map[string]interface{}) error {
	for _, v := range p.FieldList() {
		if tv, ok := targets[v.Key]; !ok {
			return fmt.Errorf("field %s is not in target", v.Key)
		} else if !reflect.DeepEqual(tv, v.Value) {
			return fmt.Errorf("field %s value is not %v, but %v", v.Key, tv, v.Value)
		}
	}

	return nil
}

func Test_GenerateInfluxPoint_Simple_Struct(t *testing.T) {
	type Data struct {
		Base      string    `influxqu:"measurement"`
		T1        string    `influxqu:"tag,t1"`
		T2        string    `influxqu:"tag,t2"`
		F1        int       `influxqu:"field,f1"`
		F2        bool      `influxqu:"field,f2"`
		Timestamp time.Time `influxqu:"timestamp"`
	}

	const (
		tag1   = "t1"
		tag2   = "t2"
		field1 = "f1"
		field2 = "f2"
	)

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

	if e := checkTags(p, map[string]string{tag1: data.T1, tag2: data.T2}); e != nil {
		t.Error(e)
	}

	if e := checkFields(p, map[string]interface{}{field1: int64(data.F1), field2: data.F2}); e != nil {
		t.Error(e)
	}
}

func Test_GenerateInfluxPoint_Custom_Type(t *testing.T) {
	type MyString string

	type Data struct {
		Base      string    `influxqu:"measurement"`
		T1        MyString  `influxqu:"tag,t1"`
		T2        string    `influxqu:"tag,t2"`
		F1        int       `influxqu:"field,f1"`
		F2        bool      `influxqu:"field,f2"`
		Timestamp time.Time `influxqu:"timestamp"`
	}

	const (
		tag1   = "t1"
		tag2   = "t2"
		field1 = "f1"
		field2 = "f2"
	)

	g := NewinfluxQu()
	data := Data{
		Base:      "base",
		T1:        MyString("t1"),
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

	if e := checkTags(p, map[string]string{tag1: string(data.T1), tag2: data.T2}); e != nil {
		t.Error(e)
	}

	if e := checkFields(p, map[string]interface{}{field1: int64(data.F1), field2: data.F2}); e != nil {
		t.Error(e)
	}
}

func Test_GenerateInfluxPoint_Nested_Struct(t *testing.T) {
	type Tag struct {
		T2 *string `influxqu:"tag,t2"`
	}

	type Field struct {
		F2 bool `influxqu:"field,f2"`
	}
	type Data struct {
		Tag
		Base      string `influxqu:"measurement"`
		T1        string `influxqu:"tag,t1"`
		F1        int    `influxqu:"field,f1"`
		F         Field
		Timestamp time.Time `influxqu:"timestamp"`
	}

	const (
		tag1   = "t1"
		tag2   = "t2"
		field1 = "f1"
		field2 = "f2"
	)

	tag2Value := "t2"

	g := NewinfluxQu()
	data := Data{
		Base:      "base",
		T1:        "t1",
		Tag:       Tag{&tag2Value},
		F1:        1,
		F:         Field{true},
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

	if e := checkTags(p, map[string]string{tag1: data.T1, tag2: *data.T2}); e != nil {
		t.Error(e)
	}

	if e := checkFields(p, map[string]interface{}{field1: int64(data.F1), field2: data.F}); e != nil {
		t.Error(e)
	}
}

func Test_GenerateInfluxPoint_Omitempty(t *testing.T) {
	type Data struct {
		Base      string    `influxqu:"measurement"`
		T1        string    `influxqu:"tag,t1,omitempty"`
		T2        string    `influxqu:"tag,t2,omitempty"`
		T3        *string   `influxqu:"tag,t3,omitempty"`
		T4        *string   `influxqu:"tag,t4,omitempty"`
		F1        int       `influxqu:"field,f1,omitempty"`
		F2        bool      `influxqu:"field,f2,omitempty"`
		F3        int32     `influxqu:"field,f3,omitempty"`
		F4        *float32  `influxqu:"field,f4,omitempty"`
		F5        *int64    `influxqu:"field,f5,omitempty"`
		Timestamp time.Time `influxqu:"timestamp"`
	}

	const (
		tag1   = "t1"
		field1 = "f1"
	)

	var (
		emptyString          = ""
		nilFloat32  *float32 = nil
	)

	g := NewinfluxQu()
	data := Data{
		Base:      "base",
		T1:        "t1",
		T2:        "",
		T3:        &emptyString,
		T4:        nil,
		F1:        1,
		F2:        false,
		F3:        0,
		F4:        nilFloat32,
		F5:        nil,
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

	if e := checkTags(p, map[string]string{tag1: data.T1}); e != nil {
		t.Error(e)
	}

	if e := checkFields(p, map[string]interface{}{field1: int64(data.F1)}); e != nil {
		t.Error(e)
	}
}

func Test_GenerateInfluxPoint_Stringer(t *testing.T) {
	type Data struct {
		Base      int       `influxqu:"measurement"`
		T1        string    `influxqu:"tag,t1"`
		T2        string    `influxqu:"tag,t2"`
		F1        int       `influxqu:"field,f1"`
		F2        bool      `influxqu:"field,f2"`
		Timestamp time.Time `influxqu:"timestamp"`
	}

	const (
		tag1   = "t1"
		tag2   = "t2"
		field1 = "f1"
		field2 = "f2"
	)

	g := NewinfluxQu()
	data := Data{
		Base:      1,
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

	if p.Name() != "1" {
		t.Error("point name is not base")
	}

	if p.Time() != data.Timestamp {
		t.Error("point timestamp is not data.Timestamp")
	}

	if e := checkTags(p, map[string]string{tag1: data.T1, tag2: data.T2}); e != nil {
		t.Error(e)
	}

	if e := checkFields(p, map[string]interface{}{field1: int64(data.F1), field2: data.F2}); e != nil {
		t.Error(e)
	}
}
