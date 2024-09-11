package internal

import (
	"sirius_future/internal/app/entity"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func DatabaseInit() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("database/test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&entity.Link{}, &entity.User{}, &entity.Payment{})
	return db
}
