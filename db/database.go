// 数据库初始化和连接

package db

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/axiaoxin-com/investool/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() {
	_, filename, _, _ := runtime.Caller(0)
	dbPath := filepath.Join(filepath.Dir(filename), "..", "data", "investool.db")

	// 创建 data 目录（如果不存在）
	dataDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		log.Printf("Failed to create data directory: %v", err)
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 检查表是否存在，如果不存在则创建
	if !DB.Migrator().HasTable(&models.Comment{}) {
		err = DB.Migrator().CreateTable(&models.Comment{})
		if err != nil {
			log.Fatalf("Failed to create table: %v", err)
		}

		// 创建索引
		err = DB.Migrator().CreateIndex(&models.Comment{}, "CreatedAt")
		if err != nil {
			log.Printf("Failed to create index on CreatedAt: %v", err)
		}

		err = DB.Migrator().CreateIndex(&models.Comment{}, "StockCode")
		if err != nil {
			log.Printf("Failed to create index on StockCode: %v", err)
		}
	}

	log.Println("Database connection established successfully")
}
