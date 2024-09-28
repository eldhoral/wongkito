package main

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/golang-module/carbon/v2"
)

func HitDigiflazz(order Order) (httpStatus int, dg ResponseDigiflazz, err error) {
	fmt.Println("Calling API...")

	reqPartnerValidity, err := json.Marshal(order)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	bodyJson := string(reqPartnerValidity)
	var ioreader io.Reader
	if bodyJson != "" {
		ioreader = bytes.NewBuffer([]byte(bodyJson))
	}
	req, err := http.NewRequest("POST", "https://api.digiflazz.com/v1/transaction", ioreader)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	var responseObject ResponseDigiflazz
	json.Unmarshal(bodyBytes, &responseObject)
	fmt.Printf("API Response as struct %+v\n", responseObject)
	dg = responseObject
	return resp.StatusCode, dg, err
}

func HitDigiflazzCekHarga() (httpStatus int, dg DigiflazzCekHargaAll, err error) {
	fmt.Println("Calling API...")

	dgCekHarga := RequestDigiflazzCekhargaALl{
		Cmd:      "prepaid",
		Username: "hubugeo6zxQo",
		Sign:     "a3df5cc72ec57829542281f12a12d071",
	}

	reqPartnerValidity, err := json.Marshal(dgCekHarga)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	bodyJson := string(reqPartnerValidity)
	var ioreader io.Reader
	if bodyJson != "" {
		ioreader = bytes.NewBuffer([]byte(bodyJson))
	}
	req, err := http.NewRequest("POST", "https://api.digiflazz.com/v1/price-list", ioreader)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	var responseObject DigiflazzCekHargaAll
	json.Unmarshal(bodyBytes, &responseObject)
	dg = responseObject
	return resp.StatusCode, dg, err
}

func orderManual(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var order Order
	_ = json.NewDecoder(r.Body).Decode(&order)
	order.Username = "hubugeo6zxQo"
	order.RefId = "manual-" + order.BuyerSkuCode + "-" + carbon.Now().ToDateMicroString()
	sign := "hubugeo6zxQoa2c8d3e0-1960-5653-8fc1-669f53cca959" + order.RefId
	data := []byte(sign)
	order.Sign = fmt.Sprintf("%x", md5.Sum(data))
	json.NewEncoder(w).Encode(&order)
}

func pembayaran(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var order Order
	_ = json.NewDecoder(r.Body).Decode(&order)
	code, response, err := pembayaranDigiflazz(order)
	if err != nil {
		var errResponse ErrorResponse
		errResponse.Status = code
		errResponse.Message = err.Error()
		errResponse.ResponseDg = response
		json.NewEncoder(w).Encode(&errResponse)
		return
	}
	if code != 200 {
		var errResponse ErrorResponse
		errResponse.Status = code
		errResponse.Message = response.Data.Message
		errResponse.ResponseDg = response
		json.NewEncoder(w).Encode(&errResponse)
		return
	}
	orders = append(orders, order)
	json.NewEncoder(w).Encode(&response)
	//go func() {
	//	err = Wr.Repository.InsertPaymentDigiflazz(response)
	//}()
}

func cekTagihan(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var cek Cek
	_ = json.NewDecoder(r.Body).Decode(&cek)
	for _, field := range orders {
		if field.RefId == cek.RefId {
			var order Order
			order.Username = field.Username
			order.BuyerSkuCode = field.BuyerSkuCode
			order.CustomerNo = field.CustomerNo
			order.RefId = field.RefId
			order.Sign = field.Sign
			code, response, err := HitDigiflazz(order)
			if err != nil {
				fmt.Print(err.Error())
				var errResponse ErrorResponse
				errResponse.Status = code
				errResponse.Message = err.Error()
				errResponse.ResponseDg = response
				json.NewEncoder(w).Encode(&errResponse)
				return
			}
			if code != 200 {
				var errResponse ErrorResponse
				errResponse.Status = code
				errResponse.Message = response.Data.Message
				errResponse.ResponseDg = response
				json.NewEncoder(w).Encode(&errResponse)
				return
			}
			json.NewEncoder(w).Encode(&response)
			return
		}
	}

}

func cekHargaDigiflazzAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	code, response, err := HitDigiflazzCekHarga()
	if err != nil {
		var errResponse ErrorResponse
		errResponse.Status = code
		errResponse.Message = err.Error()
		errResponse.ResponseDg = response
		json.NewEncoder(w).Encode(&errResponse)
		return
	}
	if code != 200 {
		var errResponse ErrorResponse
		errResponse.Status = code
		errResponse.Message = err.Error()
		errResponse.ResponseDg = response
		json.NewEncoder(w).Encode(&errResponse)
		return
	}
	json.NewEncoder(w).Encode(&response)
	go func() {
		err = Wr.Repository.DeleteAllDataDwhBillingChanneling()
		if err != nil {
			fmt.Println(err)
		}
		err = Wr.Repository.InsertAllProductDigiflazz(response)
		if err != nil {
			fmt.Println(err)
		}
	}()
	return
}

func pembayaranDigiflazz(order Order) (code int, response ResponseDigiflazz, err error) {
	var decodedByte, _ = base64.StdEncoding.DecodeString(order.PrivateKey)
	var decodedString = string(decodedByte)
	if decodedString != "wongkitostore@gmail.com" {
		return http.StatusForbidden, response, errors.New("No Authorization")
	}
	order.Username = "hubugeo6zxQo"
	order.RefId = "manual-" + order.BuyerSkuCode + "-" + carbon.Now().ToDateMicroString()
	sign := "hubugeo6zxQo8ab3cb7a-719c-4d29-b2bd-7e8ff587cf28" + order.RefId
	data := []byte(sign)
	order.Sign = fmt.Sprintf("%x", md5.Sum(data))
	code, response, err = HitDigiflazz(order)

	return
}
