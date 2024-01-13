package main

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang-module/carbon/v2"
	"github.com/gorilla/mux"
	"github.com/jasonlvhit/gocron"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cast"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Order struct {
	Username     string `json:"username"`
	BuyerSkuCode string `json:"buyer_sku_code"`
	CustomerNo   string `json:"customer_no"`
	RefId        string `json:"ref_id"`
	Sign         string `json:"sign"`
	PrivateKey   string `json:"private_key"`
}

type Cek struct {
	Username     string `json:"username"`
	BuyerSkuCode string `json:"buyer_sku_code"`
	CustomerNo   string `json:"customer_no"`
	RefId        string `json:"ref_id"`
	Sign         string `json:"sign"`
}

var orders []Order

type PesananItemkuRequest struct {
	DateStart   string `json:"date_start"`
	OrderStatus string `json:"order_status"`
}

type PesananItemkuResponse struct {
	Success bool `json:"success"`
	Data    []struct {
		OrderID             int         `json:"order_id"`
		OrderNumber         string      `json:"order_number"`
		ProductID           int         `json:"product_id"`
		Price               int         `json:"price"`
		Quantity            int         `json:"quantity"`
		GameName            string      `json:"game_name"`
		ProductName         string      `json:"product_name"`
		UsingDeliveryInfo   int         `json:"using_delivery_info"`
		DeliveryInfoField   interface{} `json:"delivery_info_field"`
		Status              string      `json:"status,omitempty"`
		RequiredInformation string      `json:"required_information"`
		DeliveryInfo        interface{} `json:"delivery_info"`
		OrderIncome         int         `json:"order_income"`
	} `json:"data"`
	Message    string `json:"message"`
	StatusCode string `json:"statusCode"`
}

type ResponseDigiflazz struct {
	Data struct {
		RefID          string `json:"ref_id"`
		CustomerNo     string `json:"customer_no"`
		BuyerSkuCode   string `json:"buyer_sku_code"`
		Message        string `json:"message"`
		Status         string `json:"status"`
		Rc             string `json:"rc"`
		BuyerLastSaldo int    `json:"buyer_last_saldo"`
		Sn             string `json:"sn"`
		Price          int    `json:"price"`
		Tele           string `json:"tele"`
		Wa             string `json:"wa"`
	} `json:"data"`
}

type ErrorResponse struct {
	Status     int         `json:"status"`
	Message    string      `json:"message"`
	ResponseDg interface{} `json:"response_digi_flazz"`
}

type DigiflazzCekHargaAll struct {
	Data []struct {
		ProductName         string `json:"product_name"`
		Category            string `json:"category"`
		Brand               string `json:"brand"`
		Type                string `json:"type"`
		SellerName          string `json:"seller_name"`
		Price               int    `json:"price"`
		BuyerSkuCode        string `json:"buyer_sku_code"`
		BuyerProductStatus  bool   `json:"buyer_product_status"`
		SellerProductStatus bool   `json:"seller_product_status"`
		UnlimitedStock      bool   `json:"unlimited_stock"`
		Stock               int    `json:"stock"`
		Multi               bool   `json:"multi"`
		StartCutOff         string `json:"start_cut_off"`
		EndCutOff           string `json:"end_cut_off"`
		Desc                string `json:"desc"`
	} `json:"data"`
}

type RequestDigiflazzCekhargaALl struct {
	Cmd      string `json:"cmd"`
	Username string `json:"username"`
	Sign     string `json:"sign"`
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
	var decodedByte, _ = base64.StdEncoding.DecodeString(order.PrivateKey)
	var decodedString = string(decodedByte)
	if decodedString != "wongkitostore@gmail.com" {
		var errResponse ErrorResponse
		errResponse.Status = http.StatusForbidden
		errResponse.Message = "Failed to process API"
		errResponse.ResponseDg = ResponseDigiflazz{}
		json.NewEncoder(w).Encode(&errResponse)
		return
	}
	order.Username = "hubugeo6zxQo"
	order.RefId = "manual-" + order.BuyerSkuCode + "-" + carbon.Now().ToDateMicroString()
	sign := "hubugeo6zxQoa2c8d3e0-1960-5653-8fc1-669f53cca959" + order.RefId
	data := []byte(sign)
	order.Sign = fmt.Sprintf("%x", md5.Sum(data))
	// foo_marshalled, err := json.Marshal(order)
	// fmt.Fprint(w, string(foo_marshalled)) // write response to ResponseWriter (w)
	code, response, err := HitDigiflazz(order)
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
	return
}

func cekPesananItemku(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	privateKey := r.Header.Get("private_key")
	var decodedByte, _ = base64.StdEncoding.DecodeString(privateKey)
	var decodedString = string(decodedByte)
	if decodedString != "wongkitostore@gmail.com" {
		var errResponse ErrorResponse
		errResponse.Status = http.StatusForbidden
		errResponse.Message = "Failed to process API"
		errResponse.ResponseDg = ResponseDigiflazz{}
		json.NewEncoder(w).Encode(&errResponse)
		return
	}

	requestPesanan := PesananItemkuRequest{DateStart: carbon.Yesterday().ToDateString(), OrderStatus: "REQUIRE_PROCESS"}
	code, response, err := hitItemkuOrderList(requestPesanan)
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
		errResponse.Message = response.Message
		errResponse.ResponseDg = response
		json.NewEncoder(w).Encode(&errResponse)
		return
	}
	if response.Success == true {
		go sendAlertToTelegram(response)
	}
	json.NewEncoder(w).Encode(&response)
	return

}

func cekPesananItemkuService() {
	requestPesanan := PesananItemkuRequest{DateStart: carbon.Yesterday().ToDateString(), OrderStatus: "REQUIRE_PROCESS"}
	code, _, err := hitItemkuOrderList(requestPesanan)
	if err != nil || code != 200 {
		SendMsgTelegram(fmt.Sprintf("Gagal cek pesanan itemku. Lihat log ini : %v || dan status code ini : %d. Segera perbaiki!!!", err, code), os.Getenv("TELEGRAM_BOT_API_KEY"), "1177093211")
	}
}

func sendAlertToTelegram(response PesananItemkuResponse) {
	numberOfOrder := len(response.Data)
	SendMsgTelegram(fmt.Sprintf("Ada %d pesanan di itemku. Segera di check!", numberOfOrder), os.Getenv("TELEGRAM_BOT_API_KEY"), "1177093211")
	for _, responseDataPesanan := range response.Data {
		reqPartnerValidity, err := json.Marshal(responseDataPesanan)
		if err != nil {
			fmt.Print(err.Error())
			return
		}
		bodyJson := string(reqPartnerValidity)
		SendMsgTelegram(strings.ReplaceAll(bodyJson, " ", "+"), os.Getenv("TELEGRAM_BOT_API_KEY"), "1177093211")
	}
}

func hitItemkuOrderList(requestItemku PesananItemkuRequest) (httpStatus int, respPI PesananItemkuResponse, err error) {
	nonce := time.Now()

	resultMarshal, err := json.Marshal(requestItemku)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	bodyJson := string(resultMarshal)
	var ioreader io.Reader
	if bodyJson != "" {
		ioreader = bytes.NewBuffer([]byte(bodyJson))
	}

	tokenBearer, err := generateJwtItemku("VFF_JLdc3Wnk5shcO3Du", cast.ToString(nonce.Unix()), requestItemku)
	req, err := http.NewRequest("POST", "https://tokoku-gateway.itemku.com/api/order/list", ioreader)
	if err != nil {
		fmt.Print(err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", "VFF_JLdc3Wnk5shcO3Du")
	req.Header.Set("Nonce", cast.ToString(nonce.Unix()))
	req.Header.Set("Authorization", "Bearer "+tokenBearer)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	defer resp.Body.Close()
	if httpStatus != 200 {
		return httpStatus, respPI, errors.New("Response status bukan 200")
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	var responseObject PesananItemkuResponse
	json.Unmarshal(bodyBytes, &responseObject)
	respPI = responseObject
	return resp.StatusCode, respPI, err
}

func generateJwtItemku(xApiKey, Nonce string, requestBody PesananItemkuRequest) (tokenString string, err error) {
	tokenJwt := jwt.Token{
		Header: map[string]interface{}{
			"X-Api-Key": xApiKey,
			"Nonce":     Nonce,
			"alg":       "HS256",
		},
		Claims: jwt.MapClaims{
			"date_start":   requestBody.DateStart,
			"order_status": requestBody.OrderStatus,
		},
		Method: jwt.SigningMethodHS256,
	}
	var secretKey = []byte("q0lo9uaXexfLZiQBkLO6iGh1G7gI-BjPrdSZeBoZ")
	tokenString, err = tokenJwt.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func SendMsgTelegram(text string, bot string, chat_id string) {

	request_url := "https://api.telegram.org/" + bot + "/sendMessage"

	client := &http.Client{}

	values := map[string]string{"text": text, "chat_id": chat_id}
	json_paramaters, _ := json.Marshal(values)

	req, _ := http.NewRequest("POST", request_url, bytes.NewBuffer(json_paramaters))
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(res.Status)
		defer res.Body.Close()
	}

}

// Main function
func main() {
	fmt.Println("mulai")

	// start assigning config url
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("Error load .env : %v", err)
		panic(err)
	}
	var configUrl ConfigEnvironment
	err = envconfig.Process("", &configUrl)
	if err != nil {
		fmt.Printf("Error assigning configUrl from .env : %v", err)
		panic(err)
	}
	gocron.Every(1).Minute().Do(cekPesananItemkuService)

	SendMsgTelegram("Backend Wongkito Jalan", os.Getenv("TELEGRAM_BOT_API_KEY"), "1177093211")

	// Init router
	r := mux.NewRouter()

	r.HandleFunc("/digiflazz/pembayaran", pembayaran).Methods("POST")
	r.HandleFunc("/digiflazz/cek", cekTagihan).Methods("POST")
	r.HandleFunc("/digiflazz/manual-pembayaran", orderManual).Methods("POST")
	r.HandleFunc("/digiflazz/cek_harga", cekHargaDigiflazzAll).Methods("GET")

	r.HandleFunc("/itemku/pesanan/cek", cekPesananItemku).Methods("POST")

	fmt.Println("API jalan")
	// Start server
	log.Fatal(http.ListenAndServe(":8090", r))

}

type ConfigEnvironment struct {
	Whatsapp Whatsapp
}
type Whatsapp struct {
	Key string `envconfig:"API_KEY_WHATSAPP" required:"true"`
}

// Request sample
// {
// 	"isbn":"4545454",
// 	"title":"Book Three",
// 	"author":{"firstname":"Harry","lastname":"White"}
// }
