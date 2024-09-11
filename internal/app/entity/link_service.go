package entity

import (
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

var (
	SecretKey = []byte("3278yd&8327dh32*(@#$E(2")
)

type Link struct {
	gorm.Model
	ID         uint   `gorm:"primaryKey"`
	Url        string `gorm:"unique;not null" json:"link"`
	ReferrerID uint   `gorm:"not null" json:"userID"`
	Count      uint   `json:"count"`
	Status     bool   `gorm:"default:true" json:"status"` // the status of the link , it can be turned off
	Limit      uint   `json:"limit"`
}

type User struct {
	gorm.Model
	ID         uint   `gorm:"primaryKey"`
	Firstname  string `gorm:"not null" json:"first_name" validate:"required,min=2,max=50"`
	Secondname string `gorm:"not null" json:"second_name" validate:"required,min=2,max=50"`
	Lastname   string `gorm:"not null" json:"last_name" validate:"required,min=2,max=50"`
	Email      string `gorm:"not null" json:"email" validate:"required,email"`
	Password   string `gorm:"not null' json:"password" validate:"required,min=2,max=50"`
	Phone      string `gorm:"not null" json:"phone" validate:"required,e164"`
	Role       string `gorm:"not null" json:"role" validate:"required"`
	ReferrerID uint   `json:"referrer_id"`
}

type Payment struct {
	ID          uint    `gorm:"primaryKey"`
	UserID      uint    `gorm:"not null" json:"user_id"`
	Amount      float64 `gorm:"not null" json:"amount"`
	Description string  `gorm:"not null" json:"description"`
	User        User    `gorm:"foreignKey:UserID"`
	Status      string  `gorm:"not null"`
}

type JWTCredentials struct {
	UserID     uint   `user_id`
	Firstname  string `json:"first_name" `
	Secondname string `json:"second_name" `
	Lastname   string `json:"last_name"`
	Email      string `json:"email"`
	Password   string `json:"password" `
	Phone      string `json:"phone"`
	Role       string `json:"role"`
	ReferrerID uint   `json:"referrer_id"`
}

var validate = validator.New()

func (u *User) Validate() error {
	return validate.Struct(u)
}
