//
// sshd.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package handlers

import (
	"fmt"
	"regexp"

	"github.com/markkurossi/lgrep/datalog"
	"github.com/markkurossi/lgrep/syslog"
)

type match struct {
	P string
	R *regexp.Regexp
}

var matches = []match{
	// Connection from 10.0.2.2 port 56821 on 10.0.2.15 port 22
	match{
		P: "sshd-connection",
		R: regexp.MustCompile(`^Connection from (\S+) port (\d+) on (\S+) port (\d+)`),
	},
	// Postponed publickey for mtr from 10.0.2.2 port 56939 ssh2 [preauth]
	match{
		P: "sshd-postponed-pubkey",
		R: regexp.MustCompile(`^Postponed publickey for (\S+) from (\S+) port (\d+) ssh2 \[preauth\]`),
	},
	// Accepted publickey for mtr from 10.0.2.2 port 56828 ssh2: RSA SHA256:R9D+G/DQmxLICfKYEoGTzKmgc48XLOa3iD6Fa4ecneY
	match{
		P: "sshd-auth-pubkey",
		R: regexp.MustCompile(`^Accepted publickey for (\S+) from (\S+) port (\d+) ssh2: (\S+) (\S+)`),
	},
	// Failed publickey for mtr from 10.0.2.2 port 56979 ssh2: RSA SHA256:R9D+G/DQmxLICfKYEoGTzKmgc48XLOa3iD6Fa4ecneY
	match{
		P: "sshd-failed-pubkey",
		R: regexp.MustCompile(`^Failed publickey for (\S+) from (\S+) port (\d+) ssh2: (\S+) (\S+)`),
	},
	// Accepted password for mtr from 10.0.2.2 port 56988 ssh2
	match{
		P: "sshd-auth-password",
		R: regexp.MustCompile(`^Accepted password for (\S+) from (\S+) port (\d+) ssh2`),
	},
	// Failed password for mtr from 10.0.2.2 port 56988 ssh2
	match{
		P: "sshd-failed-password",
		R: regexp.MustCompile(`^Failed password for (\S+) from (\S+) port (\d+) ssh2`),
	},
	// User child is on pid 4710
	match{
		P: "sshd-user-child-pid",
		R: regexp.MustCompile(`^User child is on pid (\d+)`),
	},
	// Starting session: shell on pts/8 for mtr from 10.0.2.2 port 56963 id 0
	match{
		P: "sshd-session-start",
		R: regexp.MustCompile(`^Starting session: (.*) for (\S+) from (\S+) port (\d+) id (\d+)`),
	},
	// Received disconnect from 10.0.2.2 port 56821:11: disconnected by user
	match{
		P: "sshd-disconnect",
		R: regexp.MustCompile(`^Received disconnect from (\S+) port (\d+):(.*)$`),
	},
	// Disconnected from 10.0.2.2 port 56840
	match{
		P: "sshd-disconnected",
		R: regexp.MustCompile(`^Disconnected from (\S+) port (\d+)`),
	},

	// pam_unix(sshd:session): session opened for user mtr by (uid=0)
	match{
		P: "sshd-session-open",
		R: regexp.MustCompile(`^pam_unix\(sshd:session\): session opened for user (\S+) by \(uid=(\d+)\)`),
	},
	// pam_unix(sshd:session): session closed for user mtr
	match{
		P: "sshd-session-close",
		R: regexp.MustCompile(`^pam_unix\(sshd:session\): session closed for user (\S+)`),
	},
	// pam_unix(sshd:auth): authentication failure; logname= uid=0 euid=0 tty=ssh ruser= rhost=10.0.2.2  user=mtr
	match{
		P: "sshd-authentication-failure",
		R: regexp.MustCompile(`^pam_unix\(sshd:auth\): authentication failure; logname=(\S*) uid=(\S*) euid=(\S*) tty=(\S*) ruser=(\S*) rhost=(\S*)  user=(\S*)`),
	},
}

func SSHD(e *syslog.Event, db datalog.DB) {
	for _, matcher := range matches {
		m := matcher.R.FindStringSubmatch(e.Message)
		if m == nil {
			continue
		}
		event(db, matcher.P, e, m[1:])
		return
	}
	fmt.Printf("%% SSHD: %s\n", e.Message)
}

func event(db datalog.DB, predicate string, e *syslog.Event, extra []string) {
	terms := EventTerms(e)
	for _, e := range extra {
		terms = append(terms, datalog.NewTermConstant(e, true))
	}
	sym, _ := datalog.Intern(predicate, true)
	clause := datalog.NewClause(datalog.NewAtom(sym, terms), nil)
	fmt.Printf("%s.\n", clause)
	db.Add(clause)
}
