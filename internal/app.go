package internal

import (
	"os"
	"path/filepath"
	"sirius_future/internal/app/entity"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func DatabaseInit(path string) (*gorm.DB, error) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}

	db, err := gorm.Open(sqlite.Open(path+"?_busy_timeout=5000"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)

	if err := db.AutoMigrate(&entity.Link{}, &entity.User{}, &entity.Payment{}); err != nil {
		return nil, err
	}
	return db, nil
}
