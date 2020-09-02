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

// SOAP envelope actions.
var (
	ActEnumerate       = "http://schemas.xmlsoap.org/ws/2004/09/enumeration/Enumerate"
	ActEnd             = "http://schemas.microsoft.com/wbem/wsman/1/wsman/End"
	ActHeartbeat       = "http://schemas.dmtf.org/wbem/wsman/1/wsman/Heartbeat"
	ActAck             = "http://schemas.dmtf.org/wbem/wsman/1/wsman/Ack"
	ActEvents          = "http://schemas.dmtf.org/wbem/wsman/1/wsman/Events"
	ActSubscriptionEnd = "http://schemas.xmlsoap.org/ws/2004/08/eventing/SubscriptionEnd"
)

// Envelope defines the SOAP envelope.
type Envelope struct {
	Header Header
	Body   Body
}

// AckRequested tests if the envelope requests acknowledgement.
func (e *Envelope) AckRequested() bool {
	return e.Header.AckRequested != nil
}

// Dump prints the SOAP envelope.
func (e *Envelope) Dump(label string) {
	fmt.Printf("%s\n       Action : %s\n    MessageID : %s\n  OperationID : %s\n   Identifier : %s\n AckRequested : %v\n",
		label, e.Header.Action, e.Header.MessageID, e.Header.OperationID,
		e.Header.Identifier,
		e.AckRequested())
}

// Header defines the SOAP envelope header.
type Header struct {
	Action       string
	MessageID    string
	OperationID  string
	Identifier   string
	AckRequested *AckRequested
}

// AckRequested defines the acknowledgement request status.
type AckRequested struct {
}

// Body defines the SOAP envelope body.
type Body struct {
	Events []WSManEvent `xml:"Events>Event"`
}

// WSManEvent defines a WS-Management event.
type WSManEvent struct {
	Action string `xml:"Action,attr"`
	Data   string `xml:",cdata"`
}

// Seconds define time in seconds.
type Seconds int

func (s Seconds) String() string {
	return fmt.Sprintf("PT%d.000S", s)
}

// DeliveryOptions define the event delivery options.
type DeliveryOptions struct {
	Heartbeats Seconds
	MaxTime    Seconds
}

// DeliveryNormal defines the normal delivery options.
var DeliveryNormal = &DeliveryOptions{
	Heartbeats: 900,
	MaxTime:    900,
}

// DeliveryMinLatency defines the minimum latency delivery options.
var DeliveryMinLatency = &DeliveryOptions{
	Heartbeats: 3600,
	MaxTime:    30,
}

// DeliveryMinBandwidth defines the minimum bandwidth delivery
// options.
var DeliveryMinBandwidth = &DeliveryOptions{
	Heartbeats: 21600,
	MaxTime:    21600,
}
