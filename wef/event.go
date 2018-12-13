//
// event.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package wef

import (
	"fmt"
	"strings"
)

type Event struct {
	System        System
	EventData     []EventData `xml:"EventData>Data"`
	RenderingInfo *RenderingInfo
}

func (e *Event) Dump() {
	r := &Report{}

	r.Add("Provider", e.System.Provider.Name)
	r.Add("EventID", e.System.EventID)
	r.Add("Version", e.System.Version)
	r.Add("Level", e.System.Level)
	r.Add("Task", e.System.Task)
	r.Add("Opcode", e.System.Opcode)
	r.Add("Keywords", e.System.Keywords)
	r.Add("Created", e.System.TimeCreated.SystemTime)
	r.Add("Record ID", e.System.EventRecordID)
	r.Add("Channel", e.System.Channel)
	r.Add("Computer", e.System.Computer)

	if e.System.Security != nil {
		r.Add("UserID", e.System.Security.UserID)
	}

	for _, ed := range e.EventData {
		r.Add(ed.Name, ed.Value)
	}

	if e.RenderingInfo != nil {
		r.Add("fmt.Level", e.RenderingInfo.Level)
		r.Add("fmt.Task", e.RenderingInfo.Task)
		r.Add("fmt.Opcode", e.RenderingInfo.Opcode)
		r.Add("fmt.Channel", e.RenderingInfo.Channel)
		r.Add("fmt.Provider", e.RenderingInfo.Provider)
		r.Add("fmt.Keywords", e.RenderingInfo.Keywords)
	}

	var prefix = 0
	for _, kv := range r.Data {
		if len(kv.Key) > prefix {
			prefix = len(kv.Key)
		}
	}

	for _, kv := range r.Data {
		fmt.Print(" ")
		for i := 0; i+len(kv.Key) < prefix; i++ {
			fmt.Print(" ")
		}
		fmt.Printf("%s : %s\n", kv.Key, kv.Val)
	}

	if e.RenderingInfo != nil {
		fmt.Printf("\n%s\n", e.RenderingInfo.Message)
	}
}

type Report struct {
	Data []KeyValue
}

func (r *Report) Add(key, value string) {
	value = strings.TrimSpace(value)
	if len(key) == 0 || len(value) == 0 {
		return
	}
	r.Data = append(r.Data, KeyValue{
		Key: key,
		Val: value,
	})
}

type KeyValue struct {
	Key string
	Val string
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
