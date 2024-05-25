package main

import (
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

type OrderItemkuRequest struct {
	Action       string         `json:"action"`
	DeliveryInfo []DeliveryInfo `json:"delivery_info"`
	OrderID      int            `json:"order_id"`
}
type DeliveryInfo struct {
	UsingDeliveryInfo bool    `json:"using_delivery_info"`
	DeliveryInfoField *string `json:"delivery_info_field"`
}

type DataPesananItemkuResponse struct {
	OrderID             int    `json:"order_id"`
	OrderNumber         string `json:"order_number"`
	ProductID           int    `json:"product_id"`
	Price               int    `json:"price"`
	Quantity            int    `json:"quantity"`
	GameName            string `json:"game_name"`
	ProductName         string `json:"product_name"`
	UsingDeliveryInfo   int    `json:"using_delivery_info"`
	DeliveryInfoField   string `json:"delivery_info_field"`
	Status              string `json:"status,omitempty"`
	RequiredInformation string `json:"required_information"`
	DeliveryInfo        string `json:"delivery_info"`
	OrderIncome         int    `json:"order_income"`
}

func (DataPesananItemkuResponse) TableName() string {
	return "order_itemku"
}

type PesananItemkuResponse struct {
	Success    bool                        `json:"success"`
	Data       []DataPesananItemkuResponse `json:"data"`
	Message    string                      `json:"message"`
	StatusCode string                      `json:"statusCode"`
}

type DataResponseDigiflazz struct {
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
}

func (DataResponseDigiflazz) TableName() string {
	return "payment_digiflazz"
}

type ResponseDigiflazz struct {
	Data DataResponseDigiflazz `json:"data"`
}

type ErrorResponse struct {
	Status     int         `json:"status"`
	Message    string      `json:"message"`
	ResponseDg interface{} `json:"response_digi_flazz"`
}

type DigiflazzCekHargaAll struct {
	Data []DataDigiflazzCekHargaAll `json:"data"`
}

type DataDigiflazzCekHargaAll struct {
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
}

func (DataDigiflazzCekHargaAll) TableName() string {
	return "digiflazz"
}

type RequestDigiflazzCekhargaALl struct {
	Cmd      string `json:"cmd"`
	Username string `json:"username"`
	Sign     string `json:"sign"`
}

type ConfigEnvironment struct {
	Whatsapp Whatsapp
}
type Whatsapp struct {
	Key string `envconfig:"API_KEY_WHATSAPP" required:"true"`
}

type DBMysqlOption struct {
	IsEnable             bool
	Host                 string
	Port                 int
	Username             string
	Password             string
	DBName               string
	AdditionalParameters string
	MaxOpenConns         int
	MaxIdleConns         int
	ConnMaxLifetime      time.Duration
}

type Repositories struct {
	Repository IRepository
}

type ProductItemkuDigiflazz struct {
	Id                   int    `json:"id"`
	ProductIdItemku      int    `json:"product_id_itemku"`
	ProductCodeDigiflazz string `json:"product_code_digiflazz"`
	ProductName          string `json:"product_name"`
}

func (ProductItemkuDigiflazz) TableName() string {
	return "product_itemku_digiflazz"
}
