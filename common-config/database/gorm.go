package database

import (
	"fmt"
	"strings"

	"github.com/weiqiangxu/common-config/format"

	gDriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
	gLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

const MysqlDefaultCharset = "utf8mb4"

// InitGormV2 init
func InitGormV2(cfg *format.MysqlConfig, gormLogger gLogger.Interface) (*gorm.DB, error) {
	if cfg.Charset == "" {
		cfg.Charset = MysqlDefaultCharset
	}
	dsn := fmt.Sprintf("%s:%s@"+"tcp(%s)/%s?charset=%s",
		cfg.User, cfg.Passwd, cfg.Addr, cfg.DB, cfg.Charset)
	if cfg.TimeoutSec > 0 {
		// timeout in seconds has "s"
		dsn += fmt.Sprintf("&timeout=%ds", cfg.TimeoutSec)
	}
	if !strings.Contains(cfg.Options, "parseTime=") {
		dsn += "&parseTime=true"
	}
	if !strings.Contains(cfg.Options, "loc=") {
		dsn += "&loc=Local"
	}
	// other options
	if cfg.Options != "" {
		dsn += "&" + strings.Trim(cfg.Options, "&")
	}
	db, err := gorm.Open(gDriver.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		return nil, err
	}
	// conn pool https://gorm.io/docs/connecting_to_the_database.html#Connection-Pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	// SetMaxIdleCons sets the maximum number of connections in the idle connection pool.
	if cfg.MaxIdleCount > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleCount)
	}
	// SetMaxOpenCons sets the maximum number of open connections to the database.
	if cfg.MaxOpenCount > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenCount)
	}
	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	// sqlDB.SetConnMaxLifetime(time.Hour)
	// the maximum amount of time a connection may be idle
	// SetConnMaxIdleTime; added in Go 1.15
	// sqlDB.SetConnMaxIdleTime(time.Second * 3600)
	return db, nil
}
