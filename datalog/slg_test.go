//
// slg_test.go
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

func TestSLG(t *testing.T) {
	in := strings.NewReader(input)
	parser := NewParser("data", in)

	slg := &SLG{}

	for {
		clause, clauseType, err := parser.Parse()
		if err != nil {
			if err != io.EOF {
				t.Errorf("Parser error: %s", err)
			}
			break
		}
		if false {
			fmt.Printf("%s%s\n", clause, clauseType)
		}

		switch clauseType {
		case ClauseFact:
			DBAdd(clause)

		case ClauseQuery:
			fmt.Printf("%s%s\n", clause, clauseType)
			result := slg.Query(clause.Head)
			for _, r := range result {
				fmt.Printf("=> %s\n", r)
			}
		}
	}
}
