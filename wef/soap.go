//
// soap.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package wef

import (
	"fmt"
)

var (
	ActEnumerate       = "http://schemas.xmlsoap.org/ws/2004/09/enumeration/Enumerate"
	ActEnd             = "http://schemas.microsoft.com/wbem/wsman/1/wsman/End"
	ActHeartbeat       = "http://schemas.dmtf.org/wbem/wsman/1/wsman/Heartbeat"
	ActAck             = "http://schemas.dmtf.org/wbem/wsman/1/wsman/Ack"
	ActEvents          = "http://schemas.dmtf.org/wbem/wsman/1/wsman/Events"
	ActSubscriptionEnd = "http://schemas.xmlsoap.org/ws/2004/08/eventing/SubscriptionEnd"
)

type Envelope struct {
	Header Header
	Body   Body
}

func (e *Envelope) AckRequested() bool {
	return e.Header.AckRequested != nil
}

func (e *Envelope) Dump(label string) {
	fmt.Printf("%s\n       Action : %s\n    MessageID : %s\n  OperationID : %s\n   Identifier : %s\n AckRequested : %v\n",
		label, e.Header.Action, e.Header.MessageID, e.Header.OperationID,
		e.Header.Identifier,
		e.AckRequested())
}

type Header struct {
	Action       string
	MessageID    string
	OperationID  string
	Identifier   string
	AckRequested *AckRequested
}

type AckRequested struct {
}

type Body struct {
	Events []WSManEvent `xml:"Events>Event"`
}

type WSManEvent struct {
	Action string `xml:"Action,attr"`
	Data   string `xml:",cdata"`
}
