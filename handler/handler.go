package handler

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/leopardquict/tra-statement/constant"
	"github.com/leopardquict/tra-statement/model"
)

type Handler struct {
	L *slog.Logger
}

func NewHandler(l *slog.Logger) *Handler {
	return &Handler{l}
}

func (h *Handler) GetStatement(day string, accountNumber string) ([]byte, error) {

	ref := "PBZSTM" + time.Now().Format("20060102150405")

	month := day[0:2]
	days := day[2:4]
	year := day[4:8]

	statement, err := h.GetStatementFromCbs(accountNumber, day, ref)

	if err != nil {
		h.L.Error("Error creating request", "error", err)
		return nil, err
	}

	rgs := model.Rgs{
		Statement: model.Statement{
			Header: model.Header{
				Sender:      "PBZATZTZ",
				Receiver:    "TARATZTZ",
				MsgId:       ref,
				MessageType: "STATEMENT",
			},
			MsgSummary: model.MsgSummary{
				AcctName:       statement.CustomerName,
				AcctNum:        statement.AccountNumber,
				Currency:       statement.Currency,
				CreDtTm:        time.Now().Format("2006-01-02T15:04:05"),
				SmtDt:          year + "-" + days + "-" + month,
				OpenCdtDbtInd:  "CR",
				OpenBal:        statement.OpeningBalance,
				CloseCdtDbtInd: "CR",
				CloseBal:       statement.ClosingBalance,
				NbOfTxs:        statement.TotalTransactions,
				OrgMsgId:       ref,
			},
			MessageRecords: model.MessageRecords{
				TrxRecords: []model.TrxRecord{{
					TrxDtTm:     "2021-09-01T00:00:00",
					BankRef:     "0150413430900",
					ControlNo:   "0150413430900",
					TranType:    "CR",
					TrxAmount:   "45500",
					Description: "TMS GEPG BIL 998370018394 REC 921286073786392 ERICK OYOMBE A REF FH594391634103975",
				},
					{
						TrxDtTm:     "2021-09-01T00:00:00",
						BankRef:     "0150413430900",
						ControlNo:   "0150413430900",
						TranType:    "CR",
						TrxAmount:   "45500",
						Description: "TMS GEPG BIL 998370018394 REC 921286073786392 ERICK OYOMBE A REF FH594391634103975",
					},
				},
			},
		},
		RgsSignature: "signature",
	}

	var trxRecord []model.TrxRecord
	pattern := `9\d{10,}`
	re := regexp.MustCompile(pattern)
	re2 := regexp.MustCompile(`[a-zA-Z]{1,}`)

	if len(statement.Transactions) > 0 {
		for _, transaction := range statement.Transactions {

			rec := model.TrxRecord{
				Description: transaction.Remarks,
				TrxDtTm:     transaction.TraDate + "T00:00:00",
				TranType:    transaction.DrCrDesc,
				TrxAmount:   transaction.TraAmt,
				ControlNo:   "           ",
				BankRef:     transaction.TraRefNum}

			if rec.TranType == "C" {
				rec.TranType = "CR"
			} else {
				rec.TranType = "DR"
			}

			rec.Description = ""

			match := re.FindStringSubmatch(transaction.Remarks)
			match2 := re2.FindAllString(transaction.Remarks, -1)

			if match != nil {
				if len(match) > 0 {
					rec.ControlNo = match[0]
				}
			}else{
				fmt.Println("no control number" , transaction.Remarks)
			}

			if match2 != nil {

				if len(match2) > 0 {
					for _, match := range match2 {
						rec.Description = rec.Description + " " + match
					}
				}
			}

			rec.Description = rec.Description + " " + rec.ControlNo

			trxRecord = append(trxRecord, rec)
		}
	}

	rgs.Statement.MessageRecords.TrxRecords = trxRecord

	xmlData, err := xml.Marshal(rgs.Statement)

	if err != nil {
		h.L.Error("Error marshalling statement", "error", err)
		return nil, err
	}

	absolutePath := "pbz-cer/private-key.pem"

	privateKey, err := loadPrivateKeyFromFile(absolutePath)

	if err != nil {
		fmt.Println("Error loading private key")
		h.L.Error("Error loading private key", "error", err)
	}

	// Hash the data using SHA-1
	hashed := sha1.Sum(xmlData)

	// Sign the hashed data using RSA private key
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA1, hashed[:])
	if err != nil {
		log.Fatal("Error signing the data:", err)
	}

	rgs.RgsSignature = base64.StdEncoding.EncodeToString(signature)

	Data, err := xml.Marshal(rgs)

	if err != nil {
		h.L.Error("Error marshalling statement", "error", err)
		return nil, err
	}

	//	h.L.Info("Statement retrieved request", "Statement", string(Data))

	req, err := http.NewRequest("POST", constant.URL, bytes.NewBuffer(Data))

	if err != nil {
		h.L.Error("Error creating request", "error", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("Content-type", "application/xml")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)

	if err != nil {
		h.L.Error("Error sending request", "error", err)
		return nil, err
	}

	defer resp.Body.Close()

	response, err := io.ReadAll(resp.Body)

	if err != nil {
		h.L.Error("Error reading response", "error", err)
		return nil, err

	}

	var rgsResponse model.RgsResponse

	err = xml.Unmarshal(response, &rgsResponse)

	if err != nil {
		h.L.Error("Error unmarshalling response", "error", err)
		return nil, err
	}

	h.L.Info("Response", "response", string(rgsResponse.Ack.ResponseSummary.RespStatusCode))

	if rgsResponse.Ack.ResponseSummary.RespStatusCode != "RGS001" {
		h.L.Error("Error sending statement", "error", rgsResponse.Ack.ResponseSummary.Description)
		return nil, errors.New("Error sending statement" + rgsResponse.Ack.ResponseSummary.Description)

	}

	xmlData, err = xml.Marshal(rgsResponse)

	if err != nil {
		h.L.Error("Error marshalling response", "error", err)
		return nil, err
	}

	return xmlData, nil

}

func (h *Handler) GetStatementOnDemand(day string, accountNumber string, OriginalId string) ([]byte, error) {

	ref := "PBZSTM" + time.Now().Format("20060102150405")

	month := day[0:2]
	days := day[2:4]
	year := day[4:8]

	statement, err := h.GetStatementFromCbs(accountNumber, day, ref)

	if err != nil {
		h.L.Error("Error creating request", "error", err)
		return nil, err
	}

	rgs := model.Rgs{
		Statement: model.Statement{
			Header: model.Header{
				Sender:      "PBZATZTZ",
				Receiver:    "TARATZTZ",
				MsgId:       ref,
				MessageType: "STATEMENT",
			},
			MsgSummary: model.MsgSummary{
				AcctName:       statement.CustomerName,
				AcctNum:        statement.AccountNumber,
				Currency:       statement.Currency,
				CreDtTm:        time.Now().Format("2006-01-02T15:04:05"),
				SmtDt:          year + "-" + days + "-" + month,
				OpenCdtDbtInd:  "CR",
				OpenBal:        statement.OpeningBalance,
				CloseCdtDbtInd: "CR",
				CloseBal:       statement.ClosingBalance,
				NbOfTxs:        statement.TotalTransactions,
				OrgMsgId:       OriginalId,
			},
			MessageRecords: model.MessageRecords{
				TrxRecords: []model.TrxRecord{{
					TrxDtTm:     "2021-09-01T00:00:00",
					BankRef:     "0150413430900",
					ControlNo:   "0150413430900",
					TranType:    "CR",
					TrxAmount:   "45500",
					Description: "TMS GEPG BIL 998370018394 REC 921286073786392 ERICK OYOMBE A REF FH594391634103975",
				},
					{
						TrxDtTm:     "2021-09-01T00:00:00",
						BankRef:     "0150413430900",
						ControlNo:   "0150413430900",
						TranType:    "CR",
						TrxAmount:   "45500",
						Description: "TMS GEPG BIL 998370018394 REC 921286073786392 ERICK OYOMBE A REF FH594391634103975",
					},
				},
			},
		},
		RgsSignature: "signature",
	}

	var trxRecord []model.TrxRecord
	pattern := `9\d{10,}`
	re := regexp.MustCompile(pattern)
	re2 := regexp.MustCompile(`[a-zA-Z]{1,}`)

	if len(statement.Transactions) > 0 {
		for _, transaction := range statement.Transactions {

			rec := model.TrxRecord{
				Description: transaction.Remarks,
				TrxDtTm:     transaction.TraDate + "T00:00:00",
				TranType:    transaction.DrCrDesc,
				TrxAmount:   transaction.TraAmt,
				ControlNo:   "            ",
				BankRef:     transaction.TraRefNum}

			if rec.TranType == "C" {
				rec.TranType = "CR"
			} else {
				rec.TranType = "DR"
			}

			rec.Description = ""

			match := re.FindStringSubmatch(transaction.Remarks)
			match2 := re2.FindAllString(transaction.Remarks, -1)

			if match != nil {
				if len(match) > 0 {
					rec.ControlNo = match[0]
				}
			}else{
				fmt.Println("no control number" , transaction.Remarks)
			}

			if match2 != nil {

				if len(match2) > 0 {
					for _, match := range match2 {
						rec.Description = rec.Description + " " + match
					}
				}
			}

			rec.Description = rec.Description + " " + rec.ControlNo

			trxRecord = append(trxRecord, rec)
		}
	}

	rgs.Statement.MessageRecords.TrxRecords = trxRecord

	xmlData, err := xml.Marshal(rgs.Statement)

	if err != nil {
		h.L.Error("Error marshalling statement", "error", err)
		return nil, err
	}

	absolutePath := "pbz-cer/private-key.pem"

	privateKey, err := loadPrivateKeyFromFile(absolutePath)

	if err != nil {
		fmt.Println("Error loading private key")
		h.L.Error("Error loading private key", "error", err)
	}

	// Hash the data using SHA-1
	hashed := sha1.Sum(xmlData)

	// Sign the hashed data using RSA private key
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA1, hashed[:])
	if err != nil {
		log.Fatal("Error signing the data:", err)
	}

	rgs.RgsSignature = base64.StdEncoding.EncodeToString(signature)

	Data, err := xml.Marshal(rgs)

	if err != nil {
		h.L.Error("Error marshalling statement", "error", err)
		return nil, err
	}

	//h.L.Info("Statement retrieved request", "Statement", string(Data))

	req, err := http.NewRequest("POST", constant.URL, bytes.NewBuffer(Data))

	if err != nil {
		h.L.Error("Error creating request", "error", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("Content-type", "application/xml")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)

	if err != nil {
		h.L.Error("Error sending request", "error", err)
		return nil, err
	}

	defer resp.Body.Close()

	response, err := io.ReadAll(resp.Body)

	if err != nil {
		h.L.Error("Error reading response", "error", err)
		return nil, err

	}

	h.L.Info("Response", "response", string(response))

	var rgsResponse model.RgsResponse

	err = xml.Unmarshal(response, &rgsResponse)

	if err != nil {
		h.L.Error("Error unmarshalling response", "error", err)
		return nil, err
	}

	h.L.Info("Response unmarshalled", "response", rgsResponse)

	if rgsResponse.Ack.ResponseSummary.RespStatusCode != "RGS001" {
		h.L.Error("Error sending statement", "error", rgsResponse.Ack.ResponseSummary.Description)
		return nil, errors.New("Error sending statement" + rgsResponse.Ack.ResponseSummary.Description)

	}

	xmlData, err = xml.Marshal(rgsResponse)

	if err != nil {
		h.L.Error("Error marshalling response", "error", err)
		return nil, err
	}

	return xmlData, nil

}

func loadPrivateKeyFromFile(filename string) (*rsa.PrivateKey, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read the file content
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := fileInfo.Size()
	keyBytes := make([]byte, fileSize)
	_, err = file.Read(keyBytes)
	if err != nil {
		return nil, err
	}

	// Parse the PEM block
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Parse the private key
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	// Assert the private key type to RSA
	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not an RSA key")
	}

	return rsaPrivateKey, nil
}

// verifySignature verifies the signature of the data using the public certificate

func (h *Handler) VerifySignature(data []byte, signature []byte, publicKey *rsa.PublicKey) error {
	// Hash the data using SHA-1
	hashed := sha1.Sum(data)

	// Decode the signature

	// decodedSignature, err := base64.StdEncoding.DecodeString(string(signature))
	// if err != nil {
	// 	return err
	// }

	// Verify the signature

	err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA1, hashed[:], signature)
	if err != nil {
		return err
	}

	return nil
}

// load the public key from the certificate file

func (h *Handler) loadPublicKeyFromFile(filename string) (*rsa.PublicKey, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Read the file content
	keyBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// Parse the public key in DER format
	cert, err := x509.ParseCertificate(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	// Assert the public key type to RSA
	rsaPublicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not an RSA key")
	}

	return rsaPublicKey, nil
}

func (h *Handler) StatementRequest(w http.ResponseWriter, r *http.Request) {

	var rgs model.RgsRequest

	err := xml.NewDecoder(r.Body).Decode(&rgs)

	if err != nil {
		h.L.Error("Error decoding request", "error", err)
		h.ErrorResponse(w, "400", "Error decoding request")
		return
	}

	h.L.Info("Request decoded", "request", rgs)

	if rgs.Statement.Header.MessageType == "" {
		h.L.Error("Message type missing")
		h.ErrorResponse(w, "400", "Message type missing")
		return
	}

	if rgs.Statement.Header.Sender == "" {
		h.L.Error("Sender missing")
		h.ErrorResponse(w, "400", "Sender missing")
		return

	}

	if rgs.Statement.Header.Receiver == "" {
		h.L.Error("Receiver missing")
		h.ErrorResponse(w, "400", "Receiver missing")
		return
	}

	if rgs.Statement.Header.MsgId == "" {
		h.L.Error("MsgId missing")
		h.ErrorResponse(w, "400", "MsgId missing")
		return
	}

	if rgs.Statement.RequestSummary.RequestId == "" {
		h.L.Error("RequestId missing")
		h.ErrorResponse(w, "400", "RequestId missing")
		return
	}

	if rgs.Statement.RequestSummary.CreDtTm == "" {
		h.L.Error("CreDtTm missing")
		h.ErrorResponse(w, "400", "CreDtTm missing")
		return
	}

	if rgs.Statement.RequestSummary.AcctNum == "" {
		h.L.Error("AcctNum missing")
		h.ErrorResponse(w, "400", "AcctNum missing")
		return
	}

	if rgs.Statement.RequestSummary.SmDt == "" {
		h.L.Error("SmDt missing")
		h.ErrorResponse(w, "400", "SmDt missing")
		return
	}

	if rgs.Signature == "" {
		h.L.Error("Signature missing")
		h.ErrorResponse(w, "400", "Signature missing")
		return
	}

	// absolutePath := "trauatCert.cer/trauatCert.cer"

	// publicKey, err := h.loadPublicKeyFromFile(absolutePath)

	// if err != nil {
	// 	h.L.Error("Error loading public key", "error", err)
	// 	h.ErrorResponse(w, "500", "Error loading public key")
	// 	return
	// }

	// xmlData, err := xml.Marshal(rgs.Statement)

	// if err != nil {
	// 	h.L.Error("Error marshalling statement", "error", err)
	// 	h.ErrorResponse(w, "500", "Error marshalling statement")
	// 	return
	// }

	// // Verify the signature

	// err = h.VerifySignature(xmlData, []byte(rgs.Signature), publicKey)

	// if err != nil {
	// 	h.L.Error("Error verifying signature", "error", err)
	// 	h.AckResponse(w, "REJECTED", "RGS003", "Signature Validation Failure. Signature verification failed")
	// 	return
	// }

	if constant.BANKS_ACCOUNT[rgs.Statement.RequestSummary.AcctNum] == "" {
		h.L.Error("Invalid account number")
		h.AckResponse(w, "REJECTED", "RGS004", "Invalid account number")
		return

	}

	h.AckResponse(w, "ACCEPTED", "RGS001", "Success")

	go func() {

		year := rgs.Statement.RequestSummary.SmDt[0:4]
		month := rgs.Statement.RequestSummary.SmDt[5:7]
		days := rgs.Statement.RequestSummary.SmDt[8:10]

		cobankDate := days + month + year

		_, err := h.GetStatementOnDemand(cobankDate, rgs.Statement.RequestSummary.AcctNum, rgs.Statement.Header.MsgId)

		if err != nil {
			h.L.Error("Error sending statement", "error", err)
			return
		}

		h.L.Info("Statement retrieved", "Statement", "statement generated successfully for "+rgs.Statement.RequestSummary.AcctNum+" on "+cobankDate)
	}()

}

func (h *Handler) AckResponse(w http.ResponseWriter, status string, statusCode string, description string) {
	ack := model.RgsAckRes{
		Header: model.RgsHeader{
			Sender:      "PBZATZTZ",
			Receiver:    "TARATZTZ",
			MsgId:       "PBZSTM" + time.Now().Format("20060102150405"),
			MessageType: "RESPONSE",
		},
		ResponseSummary: model.RgsResponseSummary{
			CreDtTm:        time.Now().Format("2006-01-02T15:04:05"),
			RespStatus:     status,
			RespStatusCode: statusCode,
			Description:    description,
		},
	}

	xmlData, err := xml.Marshal(ack)

	if err != nil {
		h.L.Error("Error marshalling ack", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error marshalling ack"))
		return
	}

	w.Header().Set("Content-Type", "application/xml")

	if statusCode == "RGS001" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)

	}
	w.WriteHeader(http.StatusOK)
	w.Write(xmlData)

}

func (h *Handler) ErrorResponse(w http.ResponseWriter, code string, message string) {
	errRes := model.ErrorResponse{
		Code:    code,
		Message: message,
	}

	xmlData, err := xml.Marshal(errRes)

	if err != nil {
		h.L.Error("Error marshalling error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error marshalling error"))
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(xmlData)
}
