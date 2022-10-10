// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmp

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/Jeffail/gabs"
	"github.com/golang/protobuf/proto" //nolint
	"github.com/jeremywohl/flatten"
)

// Comparer is used to compare two values. An error means the fields are either not equal or that
// an error occurred in comparing them.
type Comparer interface {
	// Eq will compare the interface values of 'a' and 'b' for equality.
	Eq(a, b interface{}) error
}

// DeepEqualComparer compares two objects using the deep equal operators.
type DeepEqualComparer struct{}

// Eq will compare using the reflect deep equal function.
func (de *DeepEqualComparer) Eq(a, b interface{}) error {
	eq := reflect.DeepEqual(a, b)
	if !eq {
		am, _ := json.Marshal(a)
		bm, _ := json.Marshal(b)
		return fmt.Errorf("%q does not exactly equal %q", string(am), string(bm))
	}
	return nil
}

// DefinedComparer tells whether one interface contains the defined fields of another. This operator relies on `omitempty` json
// tags being present in fields that shouldn't be compared when empty.
type DefinedComparer struct{}

// Eq will tell whether the fields and values defined in 'a' are present in 'b'. An error mean the values are either not equal
// or that an error occurred in comparing them.
func (dc *DefinedComparer) Eq(a, b interface{}) error {
	am, err := json.Marshal(a)
	if err != nil {
		return err
	}
	bm, err := json.Marshal(b)
	if err != nil {
		return err
	}
	bc, err := gabs.ParseJSON(bm)
	if err != nil {
		return err
	}
	aflat, err := flatten.FlattenString(string(am), "", flatten.DotStyle)
	if err != nil {
		return err
	}
	var af map[string]interface{}
	err = json.Unmarshal([]byte(aflat), &af)
	if err != nil {
		return err
	}
	contains := true
	for ak, av := range af {
		eq := reflect.DeepEqual(bc.Path(ak).Data(), av)
		if !eq {
			contains = false
			break
		}
	}
	if !contains {
		return fmt.Errorf("values in %s are not present in %s", string(am), string(bm))
	}
	return nil
}

// Contains checks if the given list contains the element using the comparer.
func Contains(list, contains interface{}, comparer Comparer) (err error) {
	switch reflect.TypeOf(list).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(list)
		for i := 0; i < s.Len(); i++ {
			o := s.Index(i)
			err = comparer.Eq(contains, o.Interface())
			if err == nil {
				return nil
			}
		}
	default:
		return fmt.Errorf("not an iterable slice: %#v", list)
	}
	return err
}

// IterableMessage is an iterable proto message type.
type IterableMessage interface {
	proto.Message
	Contains(a interface{}, comparer Comparer) error
}
