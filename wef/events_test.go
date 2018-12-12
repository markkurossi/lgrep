//
// events_test.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//
// Streaming Lossless Data Compression Algorithm - (SLDC)
// Standard ECMA-321 June 2001

package wef

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestParse(t *testing.T) {
	data, err := ioutil.ReadFile("more-events.xml")
	if err != nil {
		t.Fatalf("Failed to read input file: %s\n", err)
	}
	env := &Envelope{}
	err = xml.Unmarshal(data, env)
	if err != nil {
		t.Fatalf("Failed to parse file: %s\n", err)
	}
	for i, evt := range env.Body.Events {
		e := &Event{}
		err = xml.Unmarshal([]byte(evt.Data), e)
		if err != nil {
			t.Errorf("Failed to parse event: %s\n", err)
			continue
		}
		fmt.Printf("Event %d: %#v\n", i, e)
	}
}
