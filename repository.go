package main

import (
	"fmt"

	"gorm.io/gorm"
)

type repo struct {
	db *gorm.DB
}

func NewRepoRepository(db *gorm.DB) IRepository {
	return &repo{db: db}
}
func (r repo) InsertAllProductDigiflazz(model DigiflazzCekHargaAll) (err error) {
	err = r.db.Save(&model.Data).Error
	return
}

func (r repo) DeleteAllDataDwhBillingChanneling() (err error) {
	var data DataDigiflazzCekHargaAll
	db := r.db.Session(&gorm.Session{PrepareStmt: true})
	err = db.Exec(fmt.Sprintf("TRUNCATE TABLE %s", data.TableName())).Error
	return
}

func (r repo) InsertPaymentDigiflazz(model ResponseDigiflazz) (err error) {
	err = r.db.Save(&model.Data).Error
	return
}

func (r repo) CheckRefIdDigiflazz(model ResponseDigiflazz) (result DataResponseDigiflazz, err error) {
	err = r.db.Find(&result, "ref_id = ?", model.Data.RefID).Error
	return
}

func (r repo) InsertOrderItemku(model []DataPesananItemkuResponse) (err error) {
	err = r.db.Save(&model).Error
	return
}

func (r repo) GetMappingDataByProductIdItemku(productId int) (result ProductItemkuDigiflazz, err error) {
	err = r.db.First(&result, "product_id_itemku = ?", productId).Error
	return
}

func (r repo) FindUnprocessedOrderItemku() (result []DataPesananItemkuResponse, err error) {
	err = r.db.Find(&result, "status = ?", "REQUIRE_PROCESS").Error
	return
}

func (r repo) UpdateStatusOrderItemku(orderId []int) (rowsAffected int64, err error) {
	result := r.db.Model(DataPesananItemkuResponse{}).Where("order_id IN ?", orderId).Updates(DataPesananItemkuResponse{Status: "DELIVERED"})
	return result.RowsAffected, result.Error
}

func (r repo) GetDetailProductDigiflazz(buyerSkuCode string) (result DataDigiflazzCekHargaAll, err error) {
	err = r.db.First(&result, "buyer_sku_code = ?", buyerSkuCode).Error
	return
}
