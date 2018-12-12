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
	ActEnumerate = "http://schemas.xmlsoap.org/ws/2004/09/enumeration/Enumerate"
	ActEnd       = "http://schemas.microsoft.com/wbem/wsman/1/wsman/End"
	ActHeartbeat = "http://schemas.dmtf.org/wbem/wsman/1/wsman/Heartbeat"
	ActAck       = "http://schemas.dmtf.org/wbem/wsman/1/wsman/Ack"
)

type Envelope struct {
	Header Header
}

func (e *Envelope) Dump(label string) {
	fmt.Printf("%s\n      Action : %s\n   MessageID : %s\n OperationID : %s\n",
		label, e.Header.Action, e.Header.MessageID, e.Header.OperationID)
}

type Header struct {
	Action      string
	MessageID   string
	OperationID string
}
