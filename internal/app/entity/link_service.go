package entity

import (
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type Link struct {
	gorm.Model
	ID         uint   `gorm:"primaryKey"`
	Url        string `gorm:"unique;not null" json:"link"`
	ReferrerID uint   `gorm:"not null" json:"userID"`
	Count      uint   `json:"count"`
	Status     bool   `gorm:"default:true" json:"status"`
	Limit      uint   `json:"limit"`
}

type User struct {
	gorm.Model
	ID         uint   `gorm:"primaryKey"`
	Firstname  string `gorm:"not null" json:"first_name" validate:"required,min=2,max=50"`
	Secondname string `gorm:"not null" json:"second_name" validate:"required,min=2,max=50"`
	Lastname   string `gorm:"not null" json:"last_name" validate:"required,min=2,max=50"`
	Email      string `gorm:"not null;uniqueIndex" json:"email" validate:"required,email"`
	Password   string `gorm:"not null" json:"-" validate:"required"`
	Phone      string `gorm:"not null" json:"phone" validate:"required,e164"`
	Role       string `gorm:"not null" json:"role" validate:"required"`
	ReferrerID uint   `json:"referrer_id"`
}

type Payment struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	UserID      uint    `gorm:"not null" json:"user_id"`
	Amount      float64 `gorm:"not null" json:"amount"`
	Description string  `gorm:"not null" json:"description"`
	User        User    `gorm:"foreignKey:UserID" json:"user"`
	Status      string  `gorm:"not null" json:"status"`
}

var validate = validator.New()

func (u *User) Validate() error {
	return validate.Struct(u)
}
