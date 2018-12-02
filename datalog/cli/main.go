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

	db := datalog.NewMemDB()

	for _, arg := range flag.Args() {
		err := processFile(arg, db)
		if err != nil {
			fmt.Printf("%s\n", err)
		}
	}
}

func processFile(file string, db datalog.DB) error {
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
			db.Add(clause)

		case datalog.ClauseQuery:
			fmt.Printf("%s%s\n", clause, clauseType)
			var result []*datalog.Clause
			if false {
				result = datalog.QuerySLG(clause.Head, db, nil)
			} else {
				result = datalog.Execute(clause.Head, db, nil)
			}
			for _, r := range result {
				fmt.Printf("=> %s\n", r)
			}
		}
	}
	return nil
}
