package main

type IRepository interface {
	InsertAllProductDigiflazz(model DigiflazzCekHargaAll) (err error)
	DeleteAllDataDwhBillingChanneling() (err error)
	InsertPaymentDigiflazz(model ResponseDigiflazz) (err error)
	CheckRefIdDigiflazz(model ResponseDigiflazz) (result DataResponseDigiflazz, err error)
	InsertOrderItemku(model []DataPesananItemkuResponse) (err error)
	GetMappingDataByProductIdItemku(productId int) (result ProductItemkuDigiflazz, err error)
	FindUnprocessedOrderItemku() (result []DataPesananItemkuResponse, err error)
	UpdateStatusOrderItemku(orderId []int) (rowsAffected int64, err error)
	GetDetailProductDigiflazz(buyerSkuCode string) (result DataDigiflazzCekHargaAll, err error)
}
