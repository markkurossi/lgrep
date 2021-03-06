//
// event.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package syslog

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var reEvent = regexp.MustCompile(`^<(\d+)>([[:alpha:]]{3} [ 0-9]{2} \S+) (\S+) (.*)$`)
var reIdent = regexp.MustCompile(`^([^[]+)\[([[:digit:]]+)\]:\s*(.*)$`)

// Event implements syslog events.
type Event struct {
	Facility  Facility
	Severity  Severity
	Timestamp time.Time
	Hostname  string
	Ident     string
	Pid       int
	Message   string
}

func (e *Event) String() string {
	return fmt.Sprintf("%s:%s %s %s %s",
		e.Facility, e.Severity, e.Timestamp, e.Hostname, e.Message)
}

// Facility defines syslog event facilities.
type Facility int

// Known syslog event facilities.
const (
	Kernel Facility = iota
	UserLevel
	Mail
	System
	Security1
	Syslogd
	Printer
	News
	UUCP
	Clock1
	Security2
	FTP
	NTP
	LogAudit
	LogAlert
	Clock2
	Local0
	Local1
	Local2
	Local3
	Local4
	Local5
	Local6
	Local7
)

var facilities = map[Facility]string{
	Kernel:    "kernel",
	UserLevel: "user-level",
	Mail:      "mail",
	System:    "system",
	Security1: "security",
	Syslogd:   "syslogd",
	Printer:   "printer",
	News:      "news",
	UUCP:      "UUCP",
	Clock1:    "clock",
	Security2: "security",
	FTP:       "FTP",
	NTP:       "NTP",
	LogAudit:  "audit",
	LogAlert:  "alert",
	Clock2:    "clock",
	Local0:    "local0",
	Local1:    "local1",
	Local2:    "local2",
	Local3:    "local3",
	Local4:    "local4",
	Local5:    "local5",
	Local6:    "local6",
	Local7:    "local7",
}

func (f Facility) String() string {
	name, ok := facilities[f]
	if ok {
		return name
	}
	return fmt.Sprintf("facility_%d", f)
}

// Severity defines syslog event severities.
type Severity int

// Known syslog event severities.
const (
	Emergency Severity = iota
	Alert
	Critical
	Error
	Warning
	Notice
	Informational
	Debug
)

var severities = map[Severity]string{
	Emergency:     "emerg",
	Alert:         "alert",
	Critical:      "critical",
	Error:         "error",
	Warning:       "warning",
	Notice:        "notice",
	Informational: "info",
	Debug:         "debug",
}

func (s Severity) String() string {
	name, ok := severities[s]
	if ok {
		return name
	}
	return fmt.Sprintf("severity_%d", s)
}

// Parse parses a syslog event.
func Parse(data []byte) (*Event, error) {
	m := reEvent.FindSubmatch(data)
	if m == nil {
		return nil, fmt.Errorf("Invalid event '%s'", string(data))
	}
	priority, err := strconv.Atoi(string(m[1]))
	if err != nil {
		return nil, err
	}
	facility := priority / 8
	severity := priority % 8
	// Mon Jan 2 15:04:05 -0700 MST 2006
	timestamp, err := time.Parse("Jan _2 15:04:05", string(m[2]))
	if err != nil {
		return nil, err
	}
	now := time.Now()
	if timestamp.Month() == now.Month() {
		timestamp = timestamp.AddDate(now.Year(), 0, 0)
	}

	var ident string
	var message string
	var pid int

	mm := reIdent.FindSubmatch(m[4])
	if mm == nil {
		message = string(m[4])
	} else {
		ident = string(mm[1])
		message = string(mm[3])
		pid, err = strconv.Atoi(string(mm[2]))
		if err != nil {
			return nil, err
		}
	}

	return &Event{
		Facility:  Facility(facility),
		Severity:  Severity(severity),
		Timestamp: timestamp,
		Hostname:  string(m[3]),
		Ident:     ident,
		Pid:       pid,
		Message:   message,
	}, nil
}
