package repository

import (
	"errors"
	"sirius_future/internal/app/entity"
	"sirius_future/internal/app/service"

	"gorm.io/gorm"
)

type FutureSiriusRepository interface {
	CreateLink(link *entity.Link) error
	CreateUser(user *entity.User) error
	CheckTheLink(url string) (bool, error)

	GetAllUsers() ([]entity.User, error)

	GetAllLinks() ([]entity.Link, error)
	GetReferrerByUrl(url string) (*entity.User, error)

	CreatePayment(payment *entity.Payment) error
	GetAllPayments() ([]entity.Payment, error)
	GetPaymentsByUserID(id uint) ([]entity.Payment, error)
	UpdatePayment(id uint, payment *entity.Payment) error
}

type futureSiriusRepository struct {
	db  *gorm.DB
	log service.LoggerService
}

func NewFutureSiriusRepository(db *gorm.DB, log service.LoggerService) *futureSiriusRepository {
	return &futureSiriusRepository{db: db, log: log}
}

func (fsr *futureSiriusRepository) GetPaymentsByUserID(id uint) ([]entity.Payment, error) {
	var payments []entity.Payment
	if err := fsr.db.Preload("User").Where("user_id = ?", id).Find(&payments).Error; err != nil {
		fsr.log.Error("Error fetching payments for user ID", err, "userID", id)
		return nil, err
	}

	fsr.log.Info("Successfully fetched payments for user ID", "userID", id, "count", len(payments))
	return payments, nil
}

func (fsr *futureSiriusRepository) GetAllPayments() ([]entity.Payment, error) {
	var payments []entity.Payment
	if err := fsr.db.Preload("User").Find(&payments).Error; err != nil {
		fsr.log.Error("Error fetching all payments", err)
		return nil, err
	}

	fsr.log.Info("Successfully fetched all payments", "count", len(payments))
	return payments, nil
}

func (fsr *futureSiriusRepository) CreatePayment(payment *entity.Payment) error {
	if err := fsr.db.Create(payment).Error; err != nil {
		fsr.log.Error("Error creating payment", err, "payment", payment)
		return err
	}

	fsr.log.Info("Payment created successfully", "paymentID", payment.ID)
	return nil
}

func (fsr *futureSiriusRepository) UpdatePayment(id uint, payment *entity.Payment) error {
	var ePayment *entity.Payment
	if err := fsr.db.First(&ePayment, id).Error; err != nil {
		fsr.log.Error("Error finding payment for update", err, "paymentID", id)
		return err
	}

	// Обновляем значения только если они заданы
	if payment.Amount != 0 {
		ePayment.Amount = payment.Amount
	}
	if payment.Description != "" {
		ePayment.Description = payment.Description
	}
	if payment.Status != "" {
		ePayment.Status = payment.Status
	}

	if err := fsr.db.Save(&ePayment).Error; err != nil {
		fsr.log.Error("Error updating payment", err, "paymentID", id)
		return err
	}

	fsr.log.Info("Payment updated successfully", "paymentID", id)
	return nil
}

func (fsr *futureSiriusRepository) GetAllLinks() ([]entity.Link, error) {
	var links []entity.Link
	if err := fsr.db.Find(&links).Error; err != nil {
		fsr.log.Error("Error fetching all links", err)
		return nil, err
	}

	fsr.log.Info("Successfully fetched all links", "count", len(links))
	return links, nil
}

func (fsr *futureSiriusRepository) GetAllUsers() ([]entity.User, error) {
	var users []entity.User
	if err := fsr.db.Find(&users).Error; err != nil {
		fsr.log.Error("Error fetching all users", err)
		return nil, err
	}

	fsr.log.Info("Successfully fetched all users", "count", len(users))
	return users, nil
}

func (fsr *futureSiriusRepository) GetReferrerByUrl(url string) (*entity.User, error) {
	var link entity.Link
	if err := fsr.db.Where("url = ?", url).Find(&link).Error; err != nil {
		fsr.log.Error("Error fetching link by URL", err, "url", url)
		return nil, err
	}

	var user *entity.User
	if err := fsr.db.First(&user, link.ReferrerID).Error; err != nil {
		fsr.log.Error("Error fetching referrer by URL", err, "url", url)
		return nil, err
	}

	fsr.log.Info("Successfully fetched referrer by URL", "url", url, "referrerID", user.ID)
	return user, nil
}

func (fsr *futureSiriusRepository) CreateLink(link *entity.Link) error {
	if err := fsr.db.Create(link).Error; err != nil {
		fsr.log.Error("Error creating link", err, "link", link)
		return err
	}

	fsr.log.Info("Link created successfully", "linkID", link.ID)
	return nil
}

func (fsr *futureSiriusRepository) CreateUser(user *entity.User) error {
	if err := fsr.db.Create(user).Error; err != nil {
		fsr.log.Error("Error creating user", err, "user", user)
		return err
	}

	fsr.log.Info("User created successfully", "userID", user.ID)
	return nil
}

func (fsr *futureSiriusRepository) CheckTheLink(url string) (bool, error) {
	var link entity.Link

	result := fsr.db.Where("url = ?", url).Find(&link)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			fsr.log.Info("Link not found", "url", url)
			return false, nil
		}
		fsr.log.Error("Error fetching link by URL", result.Error, "url", url)
		return false, result.Error
	}

	if url != link.Url {
		fsr.log.Error("URL mismatch", errors.New("URL mismatch"), "providedUrl", url, "foundUrl", link.Url)
		return false, errors.New("Error")
	}

	if link.Limit <= link.Count {
		fsr.log.Info("Link usage limit reached", "url", url, "limit", link.Limit, "count", link.Count)
		return false, nil
	}

	link.Count += 1
	if err := fsr.db.Save(&link).Error; err != nil {
		fsr.log.Error("Error updating link count", err, "url", url)
		return false, err
	}

	fsr.log.Info("Link used successfully", "url", url, "newCount", link.Count)
	return true, nil
}
