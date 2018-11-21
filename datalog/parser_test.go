//
// parser_test.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package datalog

import (
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	in := strings.NewReader(input)
	parser := NewParser("data", in)
	for {
		clause, clauseType, err := parser.Parse()
		if err != nil {
			if err != io.EOF {
				t.Errorf("Parser error: %s", err)
			}
			break
		}
		fmt.Printf("%s%s\n", clause, clauseType)
	}
}
