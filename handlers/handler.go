//
// handler.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package handlers

import (
	"github.com/markkurossi/lgrep/datalog"
	"github.com/markkurossi/lgrep/syslog"
)

type Func func(e *syslog.Event, db datalog.DB, verbose bool)
