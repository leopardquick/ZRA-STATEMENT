package handler

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/leopardquict/tra-statement/constant"
	"github.com/leopardquict/tra-statement/model"
)

func CoreFunc(LL *slog.Logger) {

	h := NewHandler(LL)

	for _, v := range constant.BANKS_ACCOUNT {

		LL.Info("Processing bank: ", v, nil)
		currentTime := time.Now()

		// Subtract 24 hours to get yesterday's time
		//yesterday := currentTime.Add(-24 * time.Hour)

		// month := yesterday.Format("01")
		// day := yesterday.Format("02")
		// year := yesterday.Format("2006")

		// loadDataTomap(v, month+"-"+day+"-"+year)

		pattern := `([mtzMTZ]\d{6,7})|(2\d{8,})`

		reg := regexp.MustCompile(pattern)

		StatementResponse, err := h.GetStatementFromCbsCore(v, currentTime.Format("02012006"), currentTime.Format("02012006"), "PBZZRA"+time.Now().Format("20060102150405"))

		if err != nil {
			fmt.Println("Error in GetStatementFromCbs: ", err)
			return
		}

		if len(StatementResponse.TotalTransactions) > 0 {

			for _, transaction := range StatementResponse.Transactions[12:] {

				if constant.BANKS_ACCOUNT[transaction.TraRefNum] == "" {

					controlnumber := reg.FindString(transaction.Remarks)

					if controlnumber == "" {
						controlnumber = "0000000"
					}

					var payment model.Payment

					payment.PaymentInfo.Header.Identifier = "PBZ"
					payment.PaymentInfo.TrxInfo.BankRef = transaction.TraRefNum
					payment.PaymentInfo.TrxInfo.ControlNo = controlnumber
					payment.PaymentInfo.TrxInfo.TrxDtTm = transaction.TraDate + "T" + transaction.TraTime[:2] + ":" + transaction.TraTime[2:4] + ":" + transaction.TraTime[4:6]
					payment.PaymentInfo.TrxInfo.TrxAmount, err = strconv.ParseFloat(transaction.TraAmt, 64)

					if err != nil {
						LL.Error("Error in strconv.Atoi: ", err, payment.PaymentInfo, nil)
						continue
					}

					payment.PaymentInfo.TrxInfo.TrxCurrency = "TZS"

					if transaction.CurCode == "2" {
						payment.PaymentInfo.TrxInfo.TrxCurrency = "USD"
					}

					xmlData, err := xml.Marshal(payment.PaymentInfo)

					if err != nil {
						fmt.Println("Error in xml.Marshal: ", err)
						continue
					}

					// sign the xml

					privateKey, err := loadPrivateKeyFromFile("pbz-cert/pri.pem")

					if err != nil {
						fmt.Println("Error in loadPrivateKeyFromFile: ", err)
						continue
					}

					hashed := sha1.Sum(xmlData)
					signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA1, hashed[:])

					if err != nil {
						fmt.Println("Error in rsa.SignPKCS1v15: ", err)
						continue
					}

					payment.Signature = base64.StdEncoding.EncodeToString(signature)

					data, err := xml.Marshal(payment)

					if err != nil {
						fmt.Println("Error in xml.Marshal: ", err)
						continue
					}

					client := &http.Client{
						Timeout: time.Second * 10,
					}

					req, err := http.NewRequest("POST", "http://192.168.93.13:8080/api/v1/pbz/payment", bytes.NewBuffer(data))

					req.Header.Set("Content-Type", "application/xml")

					if err != nil {
						fmt.Println("Error in http.NewRequest: ", err)
						continue
					}

					resp, err := client.Do(req)

					if err != nil {
						fmt.Println("Error in client.Do: ", err)
						continue
					}

					defer resp.Body.Close()

					responseBody := new(bytes.Buffer)
					responseBody.ReadFrom(resp.Body)

					var response model.ResponseAck

					err = xml.Unmarshal(responseBody.Bytes(), &response)

					if err != nil {
						fmt.Println("Error in xml.Unmarshal: ", err)
						continue
					}

					LL.Info("response Body:", response.Response.RespStatus, payment.PaymentInfo.TrxInfo.BankRef, payment.PaymentInfo.TrxInfo.ControlNo, nil)

					if response.Response.RespStatusCode == "PS001" {

						constant.DATASENT[transaction.TraRefNum] = transaction.TraRefNum

					}

					time.Sleep(2 * time.Second)
				}

			}
		}

	}

}
