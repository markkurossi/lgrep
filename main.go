//
// main.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package main

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/markkurossi/lgrep/datalog"
	"github.com/markkurossi/lgrep/server"
)

func main() {
	verbose := flag.Bool("v", false, "Verbose output.")
	init := flag.String("init", "", "Init file.")
	wef := flag.String("wef", "", "Start Windows Event Forwarding server.")
	flag.Parse()

	server := server.New(datalog.NewMemDB())
	server.Verbose(*verbose)

	if len(*init) > 0 {
		err := server.Eval(*init)
		if err != nil {
			log.Fatalf("Failed to read init file: %s\n", err)
		}
	}

	if len(*wef) > 0 {
		key, err := loadKey("wef")
		if err != nil {
			log.Fatalf("Failed to load private key: %s\n", err)
		}
		cert, certBytes, err := loadCert("wef")
		if err != nil {
			log.Fatalf("Failed to load certificate: %s\n", err)
		}
		config := &tls.Config{
			Certificates: []tls.Certificate{
				tls.Certificate{
					Certificate: [][]byte{
						certBytes,
					},
					PrivateKey: key,
					Leaf:       cert,
				},
			},
			VerifyPeerCertificate: func(rawCerts [][]byte, chains [][]*x509.Certificate) error {
				fmt.Printf("chains: %v\n", chains)
				return nil
			},
			InsecureSkipVerify: true,
		}
		config.BuildNameToCertificate()
		go server.WEF.ServeHTTPS(*wef, config)
	}

	server.Syslog.ServeUDP(":1514")
}

func loadKey(path string) (*rsa.PrivateKey, error) {
	keyBytes, err := ioutil.ReadFile(fmt.Sprintf("%s.prv", path))
	if err != nil {
		return nil, err
	}
	return x509.ParsePKCS1PrivateKey(keyBytes)
}

func loadCert(path string) (*x509.Certificate, []byte, error) {
	certBytes, err := ioutil.ReadFile(fmt.Sprintf("%s.crt", path))
	if err != nil {
		return nil, nil, err
	}
	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, nil, err
	}
	return cert, certBytes, nil
}
