//
// sshd.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package syslog

import (
	"fmt"
	"regexp"

	"github.com/markkurossi/lgrep/datalog"
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
		R: regexp.MustCompile(`^Accepted publickey for (\S+) from (\S+) port (\d+) ssh2: (\S+) (\S+)$`),
	},
	// Accepted publickey for root from 10.42.0.201 port 32998 ssh2: RSA-CERT ID mtr@127.0.0.1:33872 serial 1599840225250998364 (serial 1599840225250998364) CA RSA SHA256:PADEJsxu92lFT48j4lCk1ICbaV8/hZfXQ5HAl3iTKSc
	match{
		P: "sshd-auth-certificate",
		R: regexp.MustCompile(`^Accepted publickey for (\S+) from (\S+) port (\d+) ssh2: (\S+) ID (\S+) serial (\S+) \(serial (\S+\)) CA (\S+) (\S+)`),
	},
	// Accepted certificate ID "mtr@127.0.0.1:33338 serial 8846075489776407527" (serial 8846075489776407527) signed by RSA CA SHA256:PADEJsxu92lFT48j4lCk1ICbaV8/hZfXQ5HAl3iTKSc via /etc/ssh/privx_ca.pub
	match{
		P: "sshd-accepted-certificate",
		R: regexp.MustCompile(`^Accepted certificate ID "([^"]+)" \(serial (\S+)\) signed by (\S+) CA (\S+) via (.*)`),
	},
	// error: key_cert_check_authority: invalid certificate
	match{
		P: "sshd-certificate-check-authority",
		R: regexp.MustCompile(`^error: key_cert_check_authority: (.*)$`),
	},
	// error: Certificate invalid: expired
	match{
		P: "sshd-invalid-certificate",
		R: regexp.MustCompile(`^error: Certificate invalid: (.*)$`),
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
		P: "sshd-start-session",
		R: regexp.MustCompile(`^Starting session: (.*) for (\S+) from (\S+) port (\d+) id (\d+)`),
	},
	// Close session: user mtr from 10.0.2.2 port 59132 id 0
	match{
		P: "sshd-close-session",
		R: regexp.MustCompile(`^Close session: user (\S+) from (\S+) port (\d+) id (\d+)`),
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
	// Connection closed by 10.42.0.201
	match{
		P: "sshd-connection-closed",
		R: regexp.MustCompile(`^Connection closed by (\S+)$`),
	},
	// Transferred: sent 6156, received 5544 bytes
	match{
		P: "sshd-transferred",
		R: regexp.MustCompile(`^Transferred: sent (\d+), received (\d+) bytes$`),
	},
	// Closing connection to 10.42.0.201 port 45770
	match{
		P: "sshd-closing-connection",
		R: regexp.MustCompile(`^Closing connection to (\S+) port (\d+)$`),
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
	// pam_systemd(sshd:session): Failed to release session: Interrupted system call
	match{
		P: "sshd-error-session-release",
		R: regexp.MustCompile(`^pam_systemd\(sshd:session\): Failed to release session: (.*)$`),
	},
}

func SSHD(e *Event, db datalog.DB, verbose bool) {
	for _, matcher := range matches {
		m := matcher.R.FindStringSubmatch(e.Message)
		if m == nil {
			continue
		}
		event(db, matcher.P, e, m[1:], verbose)
		return
	}
	fmt.Printf("%% SSHD: %s\n", e.Message)
}

func event(db datalog.DB, predicate string, e *Event, extra []string,
	verbose bool) {

	terms := EventTerms(e)
	for _, e := range extra {
		terms = append(terms, datalog.NewTermConstant(e, true))
	}
	sym, _ := datalog.Intern(predicate, true)
	clause := datalog.NewClause(datalog.NewAtom(sym, terms), nil)
	if verbose {
		fmt.Printf("%s.\n", clause)
	}
	db.Add(clause)
}