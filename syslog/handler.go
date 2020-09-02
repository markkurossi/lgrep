//
// handler.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package syslog

import (
	"github.com/markkurossi/datalog"
)

// Handler implements a syslog event handler.
type Handler func(e *Event, db datalog.DB, verbose bool)
