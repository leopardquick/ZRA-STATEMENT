package model

import (
	"encoding/xml"
)

type Payment struct {
	XMLName     xml.Name    `xml:"Payment"`
	PaymentInfo PaymentInfo `xml:"PaymentInfo"`
	Signature   string      `xml:"Signature"`
}

// PaymentInfo struct represents the <PaymentInfo> element
type PaymentInfo struct {
	Header  Headerser `xml:"Header"`
	TrxInfo TrxInfo   `xml:"TrxInfo"`
}

// Header struct represents the <Header> element
type Headerser struct {
	Identifier string `xml:"Identifier"`
}

// TrxInfo struct represents the <TrxInfo> element
type TrxInfo struct {
	BankRef     string  `xml:"BankRef"`
	ControlNo   string  `xml:"ControlNo"`
	TrxDtTm     string  `xml:"TrxDtTm"`
	TrxAmount   float64 `xml:"TrxAmount"`
	TrxCurrency string  `xml:"TrxCurrency"`
}

type ResponseAck struct {
	XMLName   xml.Name `xml:"ResponseAck"`
	Response  Response `xml:"Response"`
	Signature string   `xml:"Signature"`
}

// Response struct represents the <Response> element
type Response struct {
	RespStatusCode string `xml:"RespStatusCode"`
	RespStatus     string `xml:"RespStatus"`
	RspTime        string `xml:"RspTime"`
}

type PaymentReversal struct {
	ReversalInfo ReversalInfo `xml:"ReversalInfo"`
	Signature    string       `xml:"Signature"`
}

type ReversalInfo struct {
	Header  Headers  `xml:"Header"`
	TrxInfo TrxInfos `xml:"TrxInfo"`
}

type Headers struct {
	Identifier string `xml:"Identifier"`
}

type TrxInfos struct {
	BankRef     string `xml:"BankRef"`
	ControlNo   string `xml:"ControlNo"`
	RvxDtTm     string `xml:"RvxDtTm"`
	RvxAmount   int    `xml:"RvxAmount"`
	RvxCurrency string `xml:"RvxCurrency"`
}

type RequestReversalJson struct {
	BankReference    string `json:"bank_reference"`
	ControlNumber    string `json:"control_number"`
	ReversalDate     string `json:"reversal_date"`
	ReversalAmount   int    `json:"reversal_amount"`
	ReversalCurrency string `json:"reversal_currency"`
}
