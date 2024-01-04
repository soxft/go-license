package server

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

var D *gorm.DB

func loadMysql() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s", "license", "license.", "127.0.0.1", "license", "utf8mb4")

	log.Printf("trying to connect to Mysql ( %s )", dsn)

	// set logMode
	var logMode = logger.Warn

	// int sql logger
	sqlLogger := logger.New(
		log.New(os.Stderr, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Millisecond * 200, // 慢 SQL 阈值
			LogLevel:                  logMode,                // 日志级别
			IgnoreRecordNotFoundError: true,                   // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  true,                   // 禁用彩色打印
		},
	)

	// _sqlite := sqlite.Open("./tmp/test.db")

	var err error
	D, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: sqlLogger,
	})
	if err != nil {
		log.Fatalf("mysql error: %v", err)
	}

	sqlDb, err := D.DB()
	sqlDb.SetMaxOpenConns(300)
	sqlDb.SetMaxIdleConns(100)
	sqlDb.SetConnMaxLifetime(50 * time.Second)
	if err := sqlDb.Ping(); err != nil {
		log.Fatalf("mysql connect error: %v", err)
	}

	// 设置 collation
	if err := D.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci").Error; err != nil {
		log.Fatalf("mysql set collation error: %v", err)
	}

	// 自动同步表结构
	if err := D.AutoMigrate(
		License{},
	); err != nil {
		log.Fatalf("mysql auto migrate error: %v", err)
	}
}
