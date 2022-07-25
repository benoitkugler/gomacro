package test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestMarsal(t *testing.T) {
	v1 := concret2{D: 0.5}
	b, err := itfNameMarshallJSON(v1)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(b))

	v2, err := itfNameUnmarshallJSON(b)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(v1, v2) {
		t.Fatal()
	}
}

func itfNameUnmarshallJSON(src []byte) (itfName, error) {
	type wrapper struct {
		Data json.RawMessage
		Kind int
	}
	var wr wrapper
	err := json.Unmarshal(src, &wr)
	if err != nil {
		return nil, err
	}
	switch wr.Kind {
	case 0:
		var out concret1
		err = json.Unmarshal(wr.Data, &out)
		return out, err
	case 1:
		var out concret2
		err = json.Unmarshal(wr.Data, &out)
		return out, err

	default:
		panic("exhaustive switch")
	}
}

func itfNameMarshallJSON(item itfName) ([]byte, error) {
	type wrapper struct {
		Data interface{}
		Kind int
	}
	var out wrapper
	switch item.(type) {
	case concret1:
		out = wrapper{Kind: 0, Data: item}
	case concret2:
		out = wrapper{Kind: 1, Data: item}

	default:
		panic("exhaustive switch")
	}
	return json.Marshal(out)
}
