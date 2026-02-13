package storage

import (
	"log"

	storage "github.com/YnaSolyax/godrain/storage/entity"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DBStorage struct {
	db *gorm.DB
}

func NewDBStorage(dbPath string) *DBStorage {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Ошибка подключения к SQLite: %v", err)
	}

	err = db.AutoMigrate(&storage.Incident{}, &storage.Defect{}, &storage.LogItem{})
	if err != nil {
		log.Fatalf("Ошибка миграции: %v", err)
	}

	return &DBStorage{db: db}
}
