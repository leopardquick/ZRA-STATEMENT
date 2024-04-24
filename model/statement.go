package model

import "encoding/xml"

type Rgs struct {
	XMLName      xml.Name  `xml:"Rgs"`
	Statement    Statement `xml:"Statement"`
	RgsSignature string    `xml:"RgsSignature"`
}

type Statement struct {
	XMLName        xml.Name       `xml:"Statement"`
	Header         Header         `xml:"Header"`
	MsgSummary     MsgSummary     `xml:"MsgSummary"`
	MessageRecords MessageRecords `xml:"MessageRecords"`
}

type Header struct {
	XMLName     xml.Name `xml:"Header"`
	Sender      string   `xml:"Sender"`
	Receiver    string   `xml:"Receiver"`
	MsgId       string   `xml:"MsgId"`
	MessageType string   `xml:"MessageType"`
}

type MsgSummary struct {
	XMLName        xml.Name `xml:"MsgSummary"`
	AcctName       string   `xml:"AcctName"`
	AcctNum        string   `xml:"AcctNum"`
	Currency       string   `xml:"Currency"`
	CreDtTm        string   `xml:"CreDtTm"`
	SmtDt          string   `xml:"SmtDt"`
	OpenCdtDbtInd  string   `xml:"OpenCdtDbtInd"`
	OpenBal        string   `xml:"OpenBal"`
	CloseCdtDbtInd string   `xml:"CloseCdtDbtInd"`
	CloseBal       string   `xml:"CloseBal"`
	NbOfTxs        string   `xml:"NbOfTxs"`
	OrgMsgId       string   `xml:"OrgMsgId"`
}

type MessageRecords struct {
	XMLName    xml.Name    `xml:"MessageRecords"`
	TrxRecords []TrxRecord `xml:"TrxRecord"`
}

type TrxRecord struct {
	XMLName     xml.Name `xml:"TrxRecord"`
	TrxDtTm     string   `xml:"TrxDtTm"`
	BankRef     string   `xml:"BankRef"`
	ControlNo   string   `xml:"ControlNo"`
	TranType    string   `xml:"TranType"`
	TrxAmount   string   `xml:"TrxAmount"`
	Description string   `xml:"Description"`
}

type RgsResponse struct {
	Ack       RgsAck `xml:"RgsAck"`
	Signature string `xml:"RgsSignature"`
}

type RgsAck struct {
	XMLName         xml.Name        `xml:"RgsAck"`
	Header          HeaderResponse  `xml:"Header"`
	ResponseSummary ResponseSummary `xml:"ResponseSummary"`
}

type HeaderResponse struct {
	Sender      string `xml:"Sender"`
	Receiver    string `xml:"Receiver"`
	MsgId       string `xml:"MsgId"`
	MessageType string `xml:"MessageType"`
}

type ResponseSummary struct {
	CreDtTm        string `xml:"CreDtTm"`
	RespStatus     string `xml:"RespStatus"`
	RespStatusCode string `xml:"RespStatusCode"`
	Description    string `xml:"Description"`
}
