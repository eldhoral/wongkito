package main

import (
	"database/sql"
	"fmt"
	"gorm.io/gorm/logger"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func GetMysqlOptionForDWH() DBMysqlOption {
	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		dbPort = 3306
	}

	dbConnMaxOpen, err := strconv.Atoi(os.Getenv("DB_CONN_MAX_OPEN"))
	if err != nil {
		dbConnMaxOpen = 5
	}

	dbConnMaxIdle, err := strconv.Atoi(os.Getenv("DB_CONN_MAX_IDLE"))
	if err != nil {
		dbConnMaxIdle = 5
	}

	dbConnMaxLifetime, err := strconv.Atoi(os.Getenv("DB_CONN_MAX_LIFETIME"))
	if err != nil {
		dbConnMaxLifetime = 120000000000
	}
	return DBMysqlOption{
		IsEnable:             true,
		Host:                 os.Getenv("DB_HOST"),
		Port:                 dbPort,
		Username:             os.Getenv("DB_USERNAME"),
		Password:             os.Getenv("DB_PASSWORD"),
		DBName:               os.Getenv("DB_DATABASE_NAME"),
		AdditionalParameters: os.Getenv("DB_ADDITIONAL_PARAM"),
		MaxOpenConns:         dbConnMaxOpen,
		MaxIdleConns:         dbConnMaxIdle,
		ConnMaxLifetime:      time.Duration(dbConnMaxLifetime),
	}

}

// NewMysqlDatabase return gorp dbmap object with MySQL options param
func NewMysqlDatabase(option DBMysqlOption) (*gorm.DB, error) {
	dbDsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&%s", option.Username, option.Password, option.Host, option.Port, option.DBName, option.AdditionalParameters)
	timeZone := "&loc=" + strings.Replace(os.Getenv("LOG_TIME_ZONE"), "/", "%2F", -1)
	dbDsn += timeZone

	db, err := sql.Open("mysql", dbDsn)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(option.ConnMaxLifetime)
	db.SetMaxIdleConns(option.MaxIdleConns)
	db.SetMaxOpenConns(option.MaxOpenConns)

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
		NowFunc: func() time.Time {
			ti, _ := time.LoadLocation("Asia/Jakarta")
			return time.Now().In(ti)
		},
	})
	if err != nil {
		return nil, err
	}

	return gormDB, nil
}
