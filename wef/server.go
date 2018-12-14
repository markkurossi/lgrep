//
// server.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package wef

import (
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"regexp"
	"text/template"
	"unicode/utf16"

	"github.com/markkurossi/lgrep/datalog"
	"github.com/markkurossi/sldc"
)

type Server struct {
	Verbose bool
	DB      datalog.DB
}

func New(db datalog.DB) *Server {
	return &Server{
		DB: db,
	}
}

func (s *Server) ServeHTTPS(addr string, tlsConfig *tls.Config) error {

	http.HandleFunc("/wsman/SubscriptionManager/WEC",
		func(w http.ResponseWriter, r *http.Request) {
			s.subscriptionManager(w, r)
		})
	http.HandleFunc("/wsman/subscriptions/",
		func(w http.ResponseWriter, r *http.Request) {
			s.subscriptions(w, r)
		})
	http.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			dump, err := httputil.DumpRequest(r, true)
			if err != nil {
				http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
				return
			}
			fmt.Printf("Unhandled request:\n%s\n", dump)
			http.Error(w, "Not implemented yet", http.StatusNotImplemented)
		})
	httpd := &http.Server{
		Addr:      addr,
		TLSConfig: tlsConfig,
	}
	log.Printf("WEF HTTPS: listening at %s\n", addr)
	return httpd.ListenAndServeTLS("", "")
}

func (s *Server) subscriptionManager(w http.ResponseWriter, r *http.Request) {
	data, err := decodeBody(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	env := &Envelope{}
	err = xml.Unmarshal(data, env)
	if err != nil {
		log.Printf("Failed to parse enumeration request: %s\n", err)
		http.Error(w, fmt.Sprint(err), http.StatusBadRequest)
		return
	}

	switch env.Header.Action {
	case ActEnumerate:
		w.Header().Add("Content-Type", "application/soap+xml;charset=UTF-8")

		deliveryOptions := DeliveryMinLatency

		err = tmplSubscriptions.Execute(w, &Params{
			Heartbeats:       deliveryOptions.Heartbeats.String(),
			MaxTime:          deliveryOptions.MaxTime.String(),
			OperationID:      env.Header.OperationID,
			MessageID:        env.Header.MessageID,
			IssuerThumbprint: "ca5f7ce0177d3c3bf61894013af35d97caec9e40",
		})
		if err != nil {
			log.Printf("Write failed: %s\n", err)
		}

	case ActEnd:
		w.WriteHeader(http.StatusNoContent)

	default:
		fmt.Printf("Unhandled action\n")
		env.Dump(fmt.Sprintf("Subscription Manager '%s'", r.URL.Path))
		w.WriteHeader(http.StatusNoContent)
	}
}

type Params struct {
	Heartbeats       string
	MaxTime          string
	OperationID      string
	MessageID        string
	IssuerThumbprint string
}

func (s *Server) subscriptions(w http.ResponseWriter, r *http.Request) {
	dump, err := httputil.DumpRequest(r, false)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	data, err := decodeBody(r)
	if err != nil {
		fmt.Printf("Failed to decode body: %s\n", err)
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	env := &Envelope{}
	err = xml.Unmarshal(data, env)
	if err != nil {
		log.Printf("Failed to parse enumeration request: %s\n", err)
		http.Error(w, fmt.Sprint(err), http.StatusBadRequest)
		return
	}
	if s.Verbose {
		env.Dump(fmt.Sprintf("Subscription '%s'", r.URL.Path))
	}

	switch env.Header.Action {
	case ActHeartbeat, ActEnd, ActSubscriptionEnd:

	case ActEvents:
		for idx, evt := range env.Body.Events {
			e := &Event{}
			err = xml.Unmarshal([]byte(evt.Data), e)
			if err != nil {
				fmt.Printf("Failed to parse event: %s\n", err)
				continue
			}
			if s.Verbose {
				fmt.Printf("--- Event %d ----------------------------------\n",
					idx)
				e.Dump()
			}
			s.datalog(e)
		}
		s.DB.Sync()

	default:
		fmt.Printf("Unhandled action: %s\n", env.Header.Action)
		fmt.Printf("%s\n", dump)
		fmt.Printf("%s", hex.Dump(data))
	}

	if env.AckRequested() {
		w.Header().Add("Content-Type", "application/soap+xml;charset=UTF-8")
		err = tmplAck.Execute(w, &Params{
			OperationID: env.Header.OperationID,
			MessageID:   env.Header.MessageID,
		})
		if err != nil {
			log.Printf("Response write failed: %s\n", err)
		}
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

var reCharset = regexp.MustCompile("charset=([^;]+)")

func decodeBody(r *http.Request) ([]byte, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	// Content-Encoding.
	encoding := r.Header.Get("Content-Encoding")
	switch encoding {
	case "SLDC":
		decompressed, err := sldc.Decompress(body)
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			// Assume content was uncompressed.
		} else {
			body = decompressed
		}
	case "":
	default:
		return nil, fmt.Errorf("Unsupported Content-Encoding '%s'", encoding)
	}

	// Content charset.
	contentType := r.Header.Get("Content-Type")
	matches := reCharset.FindStringSubmatch(contentType)
	if matches != nil {
		switch matches[1] {
		case "UTF-16":
			body, err = decodeUTF16(body)
			if err != nil {
				return nil, err
			}
		case "UTF-8":
		default:
			return nil, fmt.Errorf("Unsupported charset '%s'", encoding)
		}
	}

	return body, nil
}

func decodeUTF16(in []byte) ([]byte, error) {
	if (len(in)%2) != 0 || len(in) < 2 {
		return nil, fmt.Errorf("Invalid UTF-16 data length")
	}
	ui16 := make([]uint16, len(in)/2-1)
	var bo binary.ByteOrder
	if in[0] == 0xff || in[1] == 0xfe {
		bo = binary.LittleEndian
	} else {
		bo = binary.BigEndian
	}

	for i := 1; i < len(in)/2; i++ {
		ui16[i-1] = bo.Uint16(in[i*2:])
	}
	return []byte(string(utf16.Decode(ui16))), nil
}

var (
	tmplSubscriptions *template.Template
	tmplAck           *template.Template
)

func init() {
	var err error
	tmplSubscriptions, err = template.New("subscriptions").Parse(`<s:Envelope
    xml:lang="en-US"
    xmlns:s="http://www.w3.org/2003/05/soap-envelope"
    xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing"
    xmlns:n="http://schemas.xmlsoap.org/ws/2004/09/enumeration"
    xmlns:w="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd"
    xmlns:p="http://schemas.microsoft.com/wbem/wsman/1/wsman.xsd">
  <s:Header>
    <a:Action>http://schemas.xmlsoap.org/ws/2004/09/enumeration/EnumerateResponse</a:Action>
    <a:MessageID>uuid:E9551077-353E-4E0B-8CA2-850776E94B09</a:MessageID>
    <a:To>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:To>
    <p:OperationID s:mustUnderstand="false">{{.OperationID}}</p:OperationID>
    <p:SequenceId>1</p:SequenceId>
    <a:RelatesTo>{{.MessageID}}</a:RelatesTo>
  </s:Header>
  <s:Body>
    <n:EnumerateResponse>
      <n:EnumerationContext>
      </n:EnumerationContext>
      <w:Items>
        <m:Subscription xmlns:m="http://schemas.microsoft.com/wbem/wsman/1/subscription">
          <m:Version>uuid:794B572A-0F96-46A6-8EFF-0F81EE927A05</m:Version>
          <s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing" xmlns:e="http://schemas.xmlsoap.org/ws/2004/08/eventing" xmlns:n="http://schemas.xmlsoap.org/ws/2004/09/enumeration" xmlns:w="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd" xmlns:p="http://schemas.microsoft.com/wbem/wsman/1/wsman.xsd">
            <s:Header>
              <a:To>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:To>
              <w:ResourceURI s:mustUnderstand="true">http://schemas.microsoft.com/wbem/wsman/1/windows/EventLog</w:ResourceURI>
              <a:ReplyTo>
                <a:Address s:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
              </a:ReplyTo>
              <a:Action s:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/08/eventing/Subscribe</a:Action>
              <w:MaxEnvelopeSize s:mustUnderstand="true">512000</w:MaxEnvelopeSize>
              <a:MessageID>uuid:290B5848-44D8-4926-B965-947E25CBD253</a:MessageID>
              <w:Locale xml:lang="en-US" s:mustUnderstand="false" />
              <p:DataLocale xml:lang="en-US" s:mustUnderstand="false" />
              <p:OperationID s:mustUnderstand="false">uuid:3D30720D-2931-4B99-A15E-C828CE3D9E8A</p:OperationID>
              <p:SequenceId s:mustUnderstand="false">1</p:SequenceId>
              <w:OptionSet xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
                <w:Option Name="SubscriptionName">Test Subscription</w:Option>
                <w:Option Name="Compression">SLDC</w:Option>
                <w:Option Name="CDATA" xsi:nil="true"/>
                <w:Option Name="ContentFormat">RenderedText</w:Option>
                <w:Option Name="IgnoreChannelError" xsi:nil="true"/>
              </w:OptionSet>
            </s:Header>
            <s:Body>
              <e:Subscribe>
                <e:EndTo>
                  <a:Address>HTTPS://10.0.2.2:15986/wsman/subscriptions/36BEB691-AFE1-458F-A6B4-8B60037F9BEE/1</a:Address>
                  <a:ReferenceProperties>
                    <e:Identifier>794B572A-0F96-46A6-8EFF-0F81EE927A05</e:Identifier>
                  </a:ReferenceProperties>
                </e:EndTo>
                <e:Delivery Mode="http://schemas.dmtf.org/wbem/wsman/1/wsman/Events">
                  <w:Heartbeats>{{.Heartbeats}}</w:Heartbeats>
                  <e:NotifyTo>
                    <a:Address>HTTPS://10.0.2.2:15986/wsman/subscriptions/36BEB691-AFE1-458F-A6B4-8B60037F9BEE/1</a:Address>
                    <a:ReferenceProperties>
                      <e:Identifier>794B572A-0F96-46A6-8EFF-0F81EE927A05</e:Identifier>
                    </a:ReferenceProperties>
                    <c:Policy xmlns:c="http://schemas.xmlsoap.org/ws/2002/12/policy" xmlns:auth="http://schemas.microsoft.com/wbem/wsman/1/authentication">
                      <c:ExactlyOne>
                        <c:All>
                          <auth:Authentication Profile="http://schemas.dmtf.org/wbem/wsman/1/wsman/secprofile/https/mutual">
                            <auth:ClientCertificate>
                              <auth:Thumbprint Role="issuer">{{.IssuerThumbprint}}</auth:Thumbprint>
                            </auth:ClientCertificate>
                          </auth:Authentication>
                        </c:All>
                      </c:ExactlyOne>
                    </c:Policy>
                  </e:NotifyTo>
                  <w:ConnectionRetry Total="5">PT60.0S</w:ConnectionRetry>
                  <w:MaxTime>{{.MaxTime}}</w:MaxTime>
                  <w:MaxEnvelopeSize Policy="Notify">512000</w:MaxEnvelopeSize>
                  <w:Locale xml:lang="en-US" s:mustUnderstand="false" />
                  <p:DataLocale xml:lang="en-US" s:mustUnderstand="false" />
                  <w:ContentEncoding>UTF-16</w:ContentEncoding>
                </e:Delivery>
                <w:Filter Dialect="http://schemas.microsoft.com/win/2004/08/events/eventquery">
                  <QueryList>
                    <Query Id="0">
                      <Select Path="Application">*[System[(Level=1  or Level=2 or Level=3 or Level=4 or Level=0 or Level=5) and TimeCreated[timediff(@SystemTime) &lt;= 86400000]]]</Select>
                      <Select Path="Security">*[System[(Level=1  or Level=2 or Level=3 or Level=4 or Level=0 or Level=5) and TimeCreated[timediff(@SystemTime) &lt;= 86400000]]]</Select>
                      <Select Path="Setup">*[System[(Level=1  or Level=2 or Level=3 or Level=4 or Level=0 or Level=5) and TimeCreated[timediff(@SystemTime) &lt;= 86400000]]]</Select>
                      <Select Path="System">*[System[(Level=1  or Level=2 or Level=3 or Level=4 or Level=0 or Level=5) and TimeCreated[timediff(@SystemTime) &lt;= 86400000]]]</Select>
                      <Select Path="ForwardedEvents">*[System[(Level=1  or Level=2 or Level=3 or Level=4 or Level=0 or Level=5) and TimeCreated[timediff(@SystemTime) &lt;= 86400000]]]</Select>
                    </Query>
                  </QueryList>
                </w:Filter>
                <w:SendBookmarks/>
              </e:Subscribe>
            </s:Body>
          </s:Envelope>
        </m:Subscription>
      </w:Items>
      <w:EndOfSequence/>
    </n:EnumerateResponse>
  </s:Body>
</s:Envelope>`)
	if err != nil {
		panic(err)
	}

	tmplAck, err = template.New("ack").Parse(`<s:Envelope
    xmlns:s="http://www.w3.org/2003/05/soap-envelope"
    xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing"
    xmlns:w="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd">
  <s:Header>
    <a:Action>http://schemas.dmtf.org/wbem/wsman/1/wsman/Ack</a:Action>
    <a:MessageID>uuid:6593DD91-ABB8-457B-AE7A-102715CAC7AC</a:MessageID>
    <a:To>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:To>
    <a:RelatesTo>{{.MessageID}}</a:RelatesTo>
  </s:Header>
  <s:Body />
</s:Envelope>`)
	if err != nil {
		panic(err)
	}
}
