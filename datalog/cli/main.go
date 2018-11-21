//
// main.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/markkurossi/lgrep/datalog"
)

func main() {
	flag.Parse()

	for _, arg := range flag.Args() {
		fmt.Printf("%s\n", arg)
		err := processFile(arg)
		if err != nil {
			fmt.Printf("%s: %s\n", arg, err)
		}
	}
}

func processFile(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	parser := datalog.NewParser(file, f)
	for {
		clause, clauseType, err := parser.Parse()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		switch clauseType {
		case datalog.ClauseFact:
			datalog.DBAdd(clause)

		case datalog.ClauseQuery:
			fmt.Printf("%s%s\n", clause, clauseType)
			result := datalog.Query(clause.Head)
			for _, r := range result {
				fmt.Printf("=> %s\n", r)
			}
		}
	}
	return nil
}
