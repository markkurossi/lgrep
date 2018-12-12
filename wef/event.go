//
// event.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package wef

type Event struct {
	System        System
	EventData     []EventData `xml:"EventData>Data"`
	RenderingInfo *RenderingInfo
}

type System struct {
	Provider      Provider
	EventID       string
	Version       string
	Level         string
	Task          string
	Opcode        string
	Keywords      string
	TimeCreated   TimeCreated
	EventRecordID string
	Correlation   *Correlation
	Execution     *Execution
	Channel       string
	Computer      string
	Security      *Security
}

type Provider struct {
	Name string `xml:"Name,attr"`
	Guid string `xml:"Guid,attr"`
}

type TimeCreated struct {
	SystemTime string `xml:",attr"`
}

type Correlation struct {
	ActivityID string `xml:",attr"`
}

type Execution struct {
	ProcessID string `xml:"ProcessID,attr"`
	ThreadID  string `xml:"ThreadID,attr"`
}

type Security struct {
	UserID string `xml:",attr"`
}

type EventData struct {
	Name  string `xml:",attr"`
	Value string `xml:",innerxml"`
}

type RenderingInfo struct {
	Message  string
	Level    string
	Task     string
	Opcode   string
	Channel  string
	Provider string
	Keywords string
}
