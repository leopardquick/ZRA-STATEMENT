package model

import "encoding/xml"

type RgsRequest struct {
	Statement StatementRequest `xml:"Statement"`
	Signature string           `xml:"RgsSignature"`
}

type StatementRequest struct {
	Header         HeaderRequest  `xml:"Header"`
	RequestSummary RequestSummary `xml:"RequestSummary"`
}

type HeaderRequest struct {
	Sender      string `xml:"Sender"`
	Receiver    string `xml:"Receiver"`
	MsgId       string `xml:"MsgId"`
	MessageType string `xml:"MessageType"`
}

type RequestSummary struct {
	RequestId string `xml:"RequestId"`
	CreDtTm   string `xml:"CreDtTm"`
	AcctNum   string `xml:"AcctNum"`
	SmDt      string `xml:"SmDt"`
}

type RgsResponseack struct {
	Ack RgsAckRes `xml:"RgsAck"`
}

type RgsAckRes struct {
	XMLName         xml.Name           `xml:"RgsAck"`
	Header          RgsHeader          `xml:"Header"`
	ResponseSummary RgsResponseSummary `xml:"ResponseSummary"`
}

type RgsHeader struct {
	Sender      string `xml:"Sender"`
	Receiver    string `xml:"Receiver"`
	MsgId       string `xml:"MsgId"`
	MessageType string `xml:"MessageType"`
}

type RgsResponseSummary struct {
	CreDtTm        string `xml:"CreDtTm"`
	RespStatus     string `xml:"RespStatus"`
	RespStatusCode string `xml:"RespStatusCode"`
	Description    string `xml:"Description"`
}

type ErrorResponse struct {
	XMLName xml.Name `xml:"Error"`
	Code    string   `xml:"Code"`
	Message string   `xml:"Message"`
}
