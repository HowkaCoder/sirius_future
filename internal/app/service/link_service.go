package service

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"sirius_future/internal/app/entity"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FutureSiriusService interface {
	CheckUserByID(id uint) error
	GenerateRefferalLink(userID uint) string
	UserValidate(user *entity.User) []string
}

type futureSiriusService struct {
	db *gorm.DB
}

func NewFutureSiriusService(db *gorm.DB) *futureSiriusService {
	return &futureSiriusService{
		db: db,
	}
}

func (fss *futureSiriusService) UserValidate(user *entity.User) []string {
	var errorMessages []string
	if err := user.Validate(); err != nil {
		validationErrors := err.(validator.ValidationErrors)

		for _, err := range validationErrors {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' Error: %s", err.Field(), err.Tag()))
		}

		return errorMessages
	}
	return errorMessages
}

func (fss *futureSiriusService) CheckUserByID(id uint) error {
	var user entity.User
	if err := fss.db.First(&user, id).Error; err != nil {
		return err
	}

	if user.ID != id {
		return errors.New("user not found")
	}
	return nil
}

func (fss *futureSiriusService) GenerateRefferalLink(userID uint) string {
	uuid := uuid.New()
	data := []byte(fmt.Sprintf("%s%d", uuid.String(), userID))

	hash := sha256.Sum256(data)

	encodedHash := base64.URLEncoding.EncodeToString(hash[:])

	return encodedHash
}
