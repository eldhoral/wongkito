package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

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

func startScheduler() {
	// set scheduler berdasarkan zona waktu sesuai kebutuhan
	jakartaTime, _ := time.LoadLocation("Asia/Jakarta")
	scheduler := cron.New(cron.WithLocation(jakartaTime))

	// stop scheduler tepat sebelum fungsi berakhir
	defer scheduler.Stop()

	scheduler.AddFunc("* * * * *", cekPesananItemkuService)
	// start scheduler
	go scheduler.Start()
}

var (
	repoos = Repositories{}
	Wr     = &repoos
)

type Student struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Error   string      `json:"error"`
	Data    interface{} `json:"data"`
}

// Main function
func main() {
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

	db, err := NewMysqlDatabase(GetMysqlOptionForDWH())
	if err != nil {
		fmt.Printf("Cannot connect to db : %v", err)
		panic(err)
	}

	Wr = wiringRepository(db)
	fmt.Println("mulai")

	startScheduler()
	SendMsgTelegram("Backend Wongkito Jalan", os.Getenv("TELEGRAM_BOT_API_KEY"), "1177093211")

	// Init router
	r := mux.NewRouter()

	r.HandleFunc("/digiflazz/pembayaran", pembayaran).Methods("POST")
	r.HandleFunc("/digiflazz/cek", cekTagihan).Methods("POST")
	r.HandleFunc("/digiflazz/manual-pembayaran", orderManual).Methods("POST")
	r.HandleFunc("/digiflazz/cek_harga", cekHargaDigiflazzAll).Methods("GET")
	r.HandleFunc("/digiflazz/pembayaran/otomatis", pembayaranOtomatis).Methods("POST")

	r.HandleFunc("/itemku/pesanan/cek", cekPesananItemku).Methods("POST")

	fmt.Println("API jalan")
	// Start server
	log.Fatal(http.ListenAndServe(":8090", r))

}

func wiringRepository(db *gorm.DB) *Repositories {
	resultRepoWiring := Repositories{
		Repository: NewRepoRepository(db),
	}
	return &resultRepoWiring
}

// Request sample
// {
// 	"isbn":"4545454",
// 	"title":"Book Three",
// 	"author":{"firstname":"Harry","lastname":"White"}
// }
