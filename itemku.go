package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang-module/carbon/v2"
	"github.com/spf13/cast"
)

func hitItemkuOrderList(requestItemku PesananItemkuRequest) (httpStatus int, respPI PesananItemkuResponse, err error) {
	nonce := time.Now()

	resultMarshal, err := json.Marshal(requestItemku)
	if err != nil {
		fmt.Println(err.Error())
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
		fmt.Println(err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", "VFF_JLdc3Wnk5shcO3Du")
	req.Header.Set("Nonce", cast.ToString(nonce.Unix()))
	req.Header.Set("Authorization", "Bearer "+tokenBearer)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
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

func generateJwtItemkuForDeliverProduct(xApiKey, Nonce string, requestBody OrderItemkuRequest) (tokenString string, err error) {
	tokenJwt := jwt.Token{
		Header: map[string]interface{}{
			"X-Api-Key": xApiKey,
			"Nonce":     Nonce,
			"alg":       "HS256",
		},
		Claims: jwt.MapClaims{
			"order_id":      requestBody.OrderID,
			"action":        requestBody.Action,
			"delivery_info": requestBody.DeliveryInfo,
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

func hitItemkuDeliverProduct(respPI PesananItemkuResponse) (err error) {
	if respPI.Success != true {
		return nil
	}
	for _, dataPesanan := range respPI.Data {
		deliveryInfoModel := DeliveryInfo{
			UsingDeliveryInfo: false,
			DeliveryInfoField: nil,
		}
		deliveryInfo := make([]DeliveryInfo, 0)
		deliveryInfo = append(deliveryInfo, deliveryInfoModel)

		request := OrderItemkuRequest{
			OrderID:      dataPesanan.OrderID,
			Action:       "DELIVER",
			DeliveryInfo: deliveryInfo,
		}

		nonce := time.Now()

		resultMarshal, err := json.Marshal(request)
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
		bodyJson := string(resultMarshal)
		var ioreader io.Reader
		if bodyJson != "" {
			ioreader = bytes.NewBuffer([]byte(bodyJson))
		}

		tokenBearer, err := generateJwtItemkuForDeliverProduct("VFF_JLdc3Wnk5shcO3Du", cast.ToString(nonce.Unix()), request)
		req, err := http.NewRequest("POST", "https://tokoku-gateway.itemku.com/api/order/action", ioreader)
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Api-Key", "VFF_JLdc3Wnk5shcO3Du")
		req.Header.Set("Nonce", cast.ToString(nonce.Unix()))
		req.Header.Set("Authorization", "Bearer "+tokenBearer)
		client := &http.Client{}
		responseItemku, err := client.Do(req)
		fmt.Println(responseItemku)
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
	}

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
		_ = hitItemkuDeliverProduct(response)
		go sendAlertToTelegram(response)
	}
	json.NewEncoder(w).Encode(&response)
	//go func() {
	//	if len(response.Data) != 0 {
	//		for i, _ := range response.Data {
	//			x := map[string]string{}
	//			err = json.Unmarshal([]byte(response.Data[i].RequiredInformation), &x)
	//			if err != nil {
	//				fmt.Println(err)
	//				continue
	//			}
	//			x["username"] = clearString(x["username"])
	//			jsonStr, err := json.Marshal(x)
	//			if err != nil {
	//				fmt.Printf("Error: %s", err.Error())
	//			}
	//			response.Data[i].RequiredInformation = string(jsonStr)
	//		}
	//		err = Wr.Repository.InsertOrderItemku(response.Data)
	//		if err != nil {
	//			fmt.Println(err)
	//			return
	//		}
	//	}
	//
	//}()
	return

}

func clearString(str string) string {
	var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
	return nonAlphanumericRegex.ReplaceAllString(str, "")
}

func cekPesananItemkuService() {
	fmt.Println(time.Now())
	fmt.Println("run cekPesananItemkuService")
	requestPesanan := PesananItemkuRequest{DateStart: carbon.Yesterday().ToDateString(), OrderStatus: "REQUIRE_PROCESS"}
	code, _, err := hitItemkuOrderList(requestPesanan)
	if err != nil || code != 200 {
		SendMsgTelegram(fmt.Sprintf("Gagal cek pesanan itemku. Lihat log ini : %v || dan status code ini : %d. Segera perbaiki!!!", err, code), os.Getenv("TELEGRAM_BOT_API_KEY"), "1177093211")
	}
}

func SchedulerPembayaran() (err error) {
	ListOrderItemku, err := Wr.Repository.FindUnprocessedOrderItemku()
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(ListOrderItemku) == 0 {
		return errors.New("Order Not Found")
	}
	OrderIdNeedToProcess := []int{}
	for _, data := range ListOrderItemku {
		if data.Status == "REQUIRE_PROCESS" {
			OrderIdNeedToProcess = append(OrderIdNeedToProcess, data.OrderID)
		}
	}
	count, err := Wr.Repository.UpdateStatusOrderItemku(OrderIdNeedToProcess)
	if err != nil {
		return
	}
	if count == int64(0) {
		return
	}

	for _, data := range ListOrderItemku {
		if data.Status == "REQUIRE_PROCESS" {

			result, err := Wr.Repository.GetMappingDataByProductIdItemku(data.ProductID)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if result.Id == 0 {
				fmt.Println(errors.New("Mapping tidak ditemukan"))
				continue
			}

			digiflazzDetailProduct, err := Wr.Repository.GetDetailProductDigiflazz(result.ProductCodeDigiflazz)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if digiflazzDetailProduct.BuyerSkuCode == "" {
				fmt.Println(errors.New("Mapping tidak ditemukan"))
				continue
			}

			if data.OrderIncome < digiflazzDetailProduct.Price {
				fmt.Println(errors.New("Modal lebih mahal untuk : " + data.OrderNumber))
				continue
			}

			//process order
			var (
				order = Order{}
				x     = map[string]string{}
			)
			err = json.Unmarshal([]byte(data.RequiredInformation), &x)
			if err != nil {
				fmt.Println(err)
				continue
			}
			customerNo := requiredInformation(data.GameName, x)
			if customerNo == "" {
				fmt.Println(errors.New("Customer Number tidak ditemukan"))
				continue
			}

			for i := 1; i <= data.Quantity; i++ {
				order = Order{
					BuyerSkuCode: result.ProductCodeDigiflazz,
					CustomerNo:   customerNo,
					PrivateKey:   "d29uZ2tpdG9zdG9yZUBnbWFpbC5jb20=",
				}

				code, response, err := pembayaranDigiflazz(order)
				if err != nil {
					fmt.Println(err)
					continue
				}
				if code != http.StatusOK {
					fmt.Println(err)
					continue
				}

				err = Wr.Repository.InsertPaymentDigiflazz(response)
				if err != nil {
					fmt.Println(err)
					continue
				}
			}

		}
	}

	return
}

func requiredInformation(gameName string, x map[string]string) (customerNo string) {
	if gameName == "Garena Free Fire" {
		customerNo = x["player_id"]
		return
	}

	if gameName == "Mobile Legends" {
		customerNo = x["player_id"] + x["zone_id"]
		return
	}
	if gameName == "Genshin Impact" {
		var zone string
		if x["zone_id"] == "Asia" {
			zone = "|os_asia"
		}
		if x["zone_id"] == "Amerika" {
			zone = "|os_usa"
		}
		if x["zone_id"] == "Europe" {
			zone = "|os_euro"
		}
		if x["zone_id"] == "TK" || x["zone_id"] == "HK" || x["zone_id"] == "MO" {
			zone = "|os_cht"
		}
		if zone == "" {
			return ""
		}
		customerNo = x["player_id"] + zone
		return
	}

	return ""
}

func pembayaranOtomatis(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := SchedulerPembayaran()
	if err != nil {
		fmt.Print(err.Error())
		var errResponse ErrorResponse
		errResponse.Status = http.StatusInternalServerError
		errResponse.Message = err.Error()
		errResponse.ResponseDg = ResponseDigiflazz{}
		json.NewEncoder(w).Encode(&errResponse)
		return
	}
	var errResponse ErrorResponse
	errResponse.Status = http.StatusOK
	errResponse.Message = "Sukses pembayaran"
	errResponse.ResponseDg = ResponseDigiflazz{}
	json.NewEncoder(w).Encode(&errResponse)
	return

}
