package test

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestMarshallJSON(t *testing.T) {
	var i union1 = member2{B: 4}
	b, err := union1Wrapper{Data: i}.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	var tmp union1Wrapper
	err = tmp.UnmarshalJSON(b)
	if err != nil {
		t.Fatal(err)
	}
	i2 := tmp.Data
	if !reflect.DeepEqual(i, i2) {
		t.Fatal(i, i2)
	}
}

func TestMarshallJSON2(t *testing.T) {
	var i union1 = member2{B: 4}
	value := StructWithITF{
		Member:        i,
		NoNeedWrapper: ITFSlice{member1(1)},
	}

	b, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}

	var value2 StructWithITF
	err = json.Unmarshal(b, &value2)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(value, value2) {
		t.Fatal(value, value2)
	}
}

func TestMarshallJSON3(t *testing.T) {
	var i union1 = member3{C: make([]string, 2)}
	b, err := union1Wrapper{Data: i}.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	var tmp union1Wrapper
	err = tmp.UnmarshalJSON(b)
	if err != nil {
		t.Fatal(err)
	}
	i2 := tmp.Data
	if !reflect.DeepEqual(i, i2) {
		t.Fatal(i, i2)
	}
}
