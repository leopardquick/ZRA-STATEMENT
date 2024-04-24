package handler

import (
	"bytes"
	"encoding/xml"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/leopardquict/tra-statement/constant"
)

type AccountStatement struct {
	XMLName xml.Name     `xml:"soap:Envelope"`
	Soap    string       `xml:"xmlns:soap,attr"`
	Ns      string       `xml:"xmlns:ns,attr"`
	Xsi     string       `xml:"xmlns:xsi,attr"`
	Xsd     string       `xml:"xmlns:xsd,attr"`
	Body    BodyStatment `xml:"soap:Body"`
}

type BodyStatment struct {
	E07Edca5 E07Edca5 `xml:"ns:e07edca5"`
}

type E07Edca5 struct {
	InpAcctType           int        `xml:"inpAcctType"`
	InpAcctKey            string     `xml:"inpAcctKey" json:"account_number"`
	InpFromDate           string     `xml:"inpFromDate" json:"from_date"`
	InpToDate             string     `xml:"inpToDate" json:"to_date"`
	ReqLanIndInout        int        `xml:"reqLanInd_inout"`
	InpChannelIdInout     int        `xml:"inpChannelId_inout"`
	InpChannelRefNumInout string     `xml:"inpChannelRefNum_inout" json:"reference_number"`
	OutAccNumOut          string     `xml:"outAccNum_out"`
	OutIbanOut            string     `xml:"outIban_out"`
	OutBbanOut            string     `xml:"outBban_out"`
	OutCusShoNameOut      string     `xml:"outCusShoName_out"`
	OutAccStaOut          int        `xml:"outAccSta_out"`
	OutAltCurOut          string     `xml:"outAltCur_out"`
	OutCurCodeOut         int        `xml:"outCurCode_out"`
	OutOpenDateOut        string     `xml:"outOpenDate_out"`
	OutCloseDateOut       string     `xml:"outCloseDate_out"`
	OutOpenBalOut         int        `xml:"outOpenBal_out"`
	OutCloseBalOut        int        `xml:"outCloseBal_out"`
	OutTotTraOut          int        `xml:"outTotTra_out"`
	OutPayloadOut         OutPayload `xml:"outPayload_out"`
	OutReqIdOut           string     `xml:"outReqId_out"`
	OutStatusOut          int        `xml:"outStatus_out"`
	OutMsgTxtOut          string     `xml:"outMsgTxt_out"`
	OutDetInfoOut         string     `xml:"outDetInfo_out"`
	DbAppErrOut           int        `xml:"dbAppErr_out"`
}

type OutPayload struct {
	XMLName xml.Name `xml:"outPayload_out"`
	Ns1     string   `xml:"xmlns:ns1,attr"`
	Xsi     string   `xml:"xsi:type,attr"`
}

type AccountStatementResponse struct {
	XMLName xml.Name         `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    BodyStatmentResp `xml:"Body"`
}

type BodyStatmentResp struct {
	Response E07edca5Response `xml:"http://com/icsfs/banks/ws/BanksMiddleware_PE07EDC00.wsdl e07edca5Response"`
}

type E07edca5Response struct {
	ReqLanIndInout        string        `xml:"reqLanInd_inout" json:"-"`
	InpChannelIdInout     string        `xml:"inpChannelId_inout" json:"-"`
	InpChannelRefNumInout string        `xml:"inpChannelRefNum_inout" json:"reference_number"`
	OutAccNumOut          string        `xml:"outAccNum_out" json:"-"`
	OutIbanOut            string        `xml:"outIban_out" json:"-"`
	OutBbanOut            string        `xml:"outBban_out" json:"account_number"`
	OutCusShoNameOut      string        `xml:"outCusShoName_out" json:"customer_name"`
	OutAccStaOut          string        `xml:"outAccSta_out" json:"account_status"`
	OutAltCurOut          string        `xml:"outAltCur_out" json:"currency"`
	OutCurCodeOut         string        `xml:"outCurCode_out" json:"-"`
	OutOpenDateOut        string        `xml:"outOpenDate_out" json:"account_open_date"`
	OutCloseDateOut       string        `xml:"outCloseDate_out" json:"-"`
	OutOpenBalOut         string        `xml:"outOpenBal_out" json:"opening_balance"`
	OutCloseBalOut        string        `xml:"outCloseBal_out" json:"closing_balance"`
	OutTotTraOut          string        `xml:"outTotTra_out" json:"total_transactions"`
	OutPayloadOut         OutPayloadOut `xml:"outPayload_out" json:"transactions"`
	OutReqIdOut           string        `xml:"outReqId_out" json:"-"`
	OutStatusOut          string        `xml:"outStatus_out" json:"-"`
	OutMsgTxtOut          string        `xml:"outMsgTxt_out" json:"message"`
	OutDetInfoOut         string        `xml:"outDetInfo_out" json:"-"`
	DbAppErrOut           string        `xml:"dbAppErr_out" json:"-"`
}

type PayloadClobObjUser struct {
	PayloadClob string `xml:"payloadclob" json:"transaction_details"`
	SeqNum      int    `xml:"seqNum" json:"seq_num"`
}

type OutPayloadOut struct {
	PayloadClobObjUser []PayloadClobObjUser `xml:"PayloadClobObjUser" json:"transaction"`
}

func NewAccountStatement() *AccountStatement {
	return &AccountStatement{
		Soap: "http://schemas.xmlsoap.org/soap/envelope/",
		Ns:   "http://com/icsfs/banks/ws/BanksMiddleware_PE07EDC00.wsdl",
		Xsi:  "http://www.w3.org/2001/XMLSchema-instance",
		Xsd:  "http://www.w3.org/2001/XMLSchema",
		Body: BodyStatment{
			E07Edca5: E07Edca5{

				OutPayloadOut: OutPayload{
					Ns1: "http://com/icsfs/banks/ws/BanksMiddleware_PE07EDC00.wsdl/types/",
					Xsi: "ns1:PayloadClobObjUserArray",
				},
			},
		},
	}
}

type StatementResponse struct {
	ReferenceNumber   string        `json:"reference_number"`
	AccountNumber     string        `json:"account_number"`
	CustomerName      string        `json:"customer_name"`
	AccountStatus     string        `json:"account_status"`
	Currency          string        `json:"currency"`
	OpeningBalance    string        `json:"opening_balance"`
	ClosingBalance    string        `json:"closing_balance"`
	TotalTransactions string        `json:"total_transactions"`
	Transactions      []Transaction `json:"transactions"`
	Message           string        `json:"message"`
}

type Transaction struct {
	BRACode      string `xml:"BRA_CODE" json:"-"`
	CustomerNum  string `xml:"CUS_NUM" json:"customer_number"`
	CurCode      string `xml:"CUR_CODE"  `
	LedCode      string `xml:"LED_CODE"`
	SubAcctCode  string `xml:"SUB_ACCT_CODE"`
	TraDate      string `xml:"TRA_DATE" json:"transaction_date"`
	ActTraDate   string `xml:"ACT_TRA_DATE" json:"actual_transaction_date"`
	TraTime      string `xml:"TRA_TIME" json:"transaction_time"`
	TraSeq1      string `xml:"TRA_SEQ1" json:"-"`
	TraSeq2      string `xml:"TRA_SEQ2" json:"-"`
	TraAmt       string `xml:"TRA_AMT" json:"transaction_amount"`
	EquTraAmt    string `xml:"EQU_TRA_AMT" json:"equivalent_transaction_amount"`
	CurPri       string `xml:"CUR_PRI" json:"currency_price"`
	IntDate      string `xml:"INT_DATE" json:"-"`
	ValDate      string `xml:"VAL_DATE" json:"value_date"`
	DrCrInd      string `xml:"DR_CR_IND" json:"dr_cr_ind"`
	DrCrDesc     string `xml:"DR_CR_DESC" json:"dr_cr_desc"`
	CrntBal      string `xml:"CRNT_BAL" json:"current_balance"`
	DocNum       string `xml:"DOC_NUM" json:"document_number"`
	DocAlp       string `xml:"DOC_ALP" json:"-"`
	Remarks      string `xml:"REMARKS" json:"remarks"`
	ExplCode     string `xml:"EXPL_CODE" json:"expl_code"`
	ExplEng      string `xml:"EXPL_ENG" json:"expl_eng"`
	ExplNat      string `xml:"EXPL_NAT" json:"-"`
	OrigTellID   string `xml:"ORIG_TELL_ID" json:"-"`
	OrigTellDesc string `xml:"ORIG_TELL_DESC" json:"-"`
	OrigBraCode  string `xml:"ORIG_BRA_CODE" json:"-"`
	OrigTraDate  string `xml:"ORIG_TRA_DATE" json:"-"`
	OrigTraSeq1  string `xml:"ORIG_TRA_SEQ1" json:"-"`
	OrigTraSeq2  string `xml:"ORIG_TRA_SEQ2" json:"-"`
	TraRefNum    string `xml:"TRA_REF_NUM" json:"transaction_reference_number"`
}

type AccountStatementRequest struct {
	ReferenceNumber string `json:"reference_number"`
	AccountNumber   string `json:"account_number"`
	FromDate        string `json:"from_date"`
	ToDate          string `json:"to_date"`
}

func (h *Handler) GetStatementFromCbs(accountNumber string, fromDate string, referenceNumber string) (StatementResponse, error) {

	// get the channel id from the context

	channelID := "34"

	var accountStatementRequest AccountStatementRequest

	// Decode the incoming AccountStatementRequest json

	accountStatementRequest.AccountNumber = accountNumber
	accountStatementRequest.FromDate = fromDate
	accountStatementRequest.ToDate = fromDate
	accountStatementRequest.ReferenceNumber = "PBZSTM" + time.Now().Format("20060102150405")

	// check if account number is empty

	r := NewAccountStatement()

	r.Body.E07Edca5.InpAcctKey = accountStatementRequest.AccountNumber
	r.Body.E07Edca5.InpChannelRefNumInout = accountStatementRequest.ReferenceNumber
	r.Body.E07Edca5.InpFromDate = accountStatementRequest.FromDate
	r.Body.E07Edca5.InpToDate = accountStatementRequest.ToDate

	// check if reference number is empty

	r.Body.E07Edca5.InpAcctType = 2
	r.Body.E07Edca5.ReqLanIndInout = 1

	channelIDInt, err := strconv.Atoi(channelID)

	if err != nil {
		h.L.Error("fail to mashal: %v", err.Error(), nil)
		return StatementResponse{}, err
	}

	r.Body.E07Edca5.InpChannelIdInout = channelIDInt

	xmlData, err := xml.Marshal(r)

	if err != nil {
		h.L.Error("fail to mashal", err.Error(), nil)
		return StatementResponse{}, err
	}

	req, err := http.NewRequest("POST", constant.ACCOUNT_STATEMENT_URL, bytes.NewBuffer(xmlData))

	if err != nil {
		h.L.Error("error %v", err.Error(), nil)
		return StatementResponse{}, err
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "http://com/icsfs/banks/ws/BanksMiddleware_PE07EDC00.wsdl ")

	// Perform the HTTP request
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		h.L.Error("error %v", err.Error(), nil)
		return StatementResponse{}, err
	}

	defer resp.Body.Close()

	// write the response to the response writer

	responseBody := new(bytes.Buffer)
	responseBody.ReadFrom(resp.Body)

	var responseEnvelope AccountStatementResponse

	err = xml.Unmarshal(responseBody.Bytes(), &responseEnvelope)

	if err != nil {
		h.L.Error("error %v", err.Error(), nil)
		return StatementResponse{}, err
	}

	if responseEnvelope.Body.Response.OutStatusOut != "0" {
		h.L.Error("error %v", responseEnvelope.Body.Response.OutMsgTxtOut, nil)
		return StatementResponse{}, errors.New(responseEnvelope.Body.Response.OutMsgTxtOut)
	}

	if responseEnvelope.Body.Response.OutAccStaOut == "1" {
		responseEnvelope.Body.Response.OutAccStaOut = "open"
	} else if responseEnvelope.Body.Response.OutAccStaOut == "2" {
		responseEnvelope.Body.Response.OutAccStaOut = "closed"
	} else if responseEnvelope.Body.Response.OutAccStaOut == "3" {
		responseEnvelope.Body.Response.OutAccStaOut = "dormant"
	}

	if responseEnvelope.Body.Response.OutStatusOut == "0" {
		responseEnvelope.Body.Response.OutStatusOut = "success"
	} else if responseEnvelope.Body.Response.OutStatusOut == "-13" {
		responseEnvelope.Body.Response.OutStatusOut = "duplicated reference number"

		h.L.Error("error %v", responseEnvelope.Body.Response.OutStatusOut, nil)

		// new error

		return StatementResponse{}, errors.New(responseEnvelope.Body.Response.OutStatusOut)
	}

	var allTransactions []Transaction

	for _, transaction := range responseEnvelope.Body.Response.OutPayloadOut.PayloadClobObjUser {
		transaction, err := r.ParseXMLToTransaction(transaction.PayloadClob)

		if err != nil {
			h.L.Error("error %v", responseEnvelope.Body.Response.OutMsgTxtOut, nil)
			return StatementResponse{}, errors.New(responseEnvelope.Body.Response.OutMsgTxtOut)
		}

		allTransactions = append(allTransactions, *transaction)
	}

	statementResponse := StatementResponse{
		ReferenceNumber:   responseEnvelope.Body.Response.InpChannelRefNumInout,
		AccountNumber:     responseEnvelope.Body.Response.OutBbanOut,
		CustomerName:      responseEnvelope.Body.Response.OutCusShoNameOut,
		AccountStatus:     responseEnvelope.Body.Response.OutAccStaOut,
		Currency:          responseEnvelope.Body.Response.OutAltCurOut,
		OpeningBalance:    responseEnvelope.Body.Response.OutOpenBalOut,
		ClosingBalance:    responseEnvelope.Body.Response.OutCloseBalOut,
		TotalTransactions: responseEnvelope.Body.Response.OutTotTraOut,
		Transactions:      allTransactions,
		Message:           responseEnvelope.Body.Response.OutMsgTxtOut,
	}

	// write the response to the response writer

	return statementResponse, nil

}

func (r *AccountStatement) ParseXMLToTransaction(xmlData string) (*Transaction, error) {
	// Create a decoder and use it to decode the XML string
	decoder := xml.NewDecoder(strings.NewReader("<root>" + xmlData + "</root>"))

	// Create a variable of the struct type to store the decoded data
	var transaction Transaction

	// Loop through the XML elements and decode them into the struct
	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch se := token.(type) {
		case xml.StartElement:
			// Decode XML element and assign it to the corresponding field in the struct
			if err := decoder.DecodeElement(&transaction, &se); err != nil {
				return nil, err
			}
		}
	}

	return &transaction, nil
}
