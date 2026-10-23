package repository

import (
	"errors"
	"sirius_future/internal/app/entity"
	"sirius_future/internal/app/service"

	"gorm.io/gorm"
)

var (
	ErrLinkNotFound  = errors.New("link not found")
	ErrLinkExhausted = errors.New("link is disabled or its limit is reached")
)

type FutureSiriusRepository interface {
	CreateLink(link *entity.Link) error
	CreateUser(user *entity.User) error
	CheckTheLink(url string) (bool, error)
	CreateUserWithReferral(user *entity.User, url string) error

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

	payment.UserID = ePayment.UserID

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
	if err := fsr.db.Where("url = ?", url).First(&link).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLinkNotFound
		}
		fsr.log.Error("Error fetching link by URL", err, "url", url)
		return nil, err
	}

	var user entity.User
	if err := fsr.db.First(&user, link.ReferrerID).Error; err != nil {
		fsr.log.Error("Error fetching referrer by URL", err, "url", url)
		return nil, err
	}
	return &user, nil
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
	if err := fsr.db.Where("url = ?", url).First(&link).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, ErrLinkNotFound
		}
		fsr.log.Error("Error fetching link by URL", err, "url", url)
		return false, err
	}
	return link.Status && link.Count < link.Limit, nil
}

func (fsr *futureSiriusRepository) CreateUserWithReferral(user *entity.User, url string) error {
	return fsr.db.Transaction(func(tx *gorm.DB) error {
		var link entity.Link
		if err := tx.Where("url = ?", url).First(&link).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrLinkNotFound
			}
			return err
		}

		ok, err := useLink(tx, url)
		if err != nil {
			return err
		}
		if !ok {
			return ErrLinkExhausted
		}

		user.ReferrerID = link.ReferrerID
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		fsr.log.Info("User registered by referral link", "userID", user.ID, "referrerID", link.ReferrerID)
		return nil
	})
}

func useLink(db *gorm.DB, url string) (bool, error) {
	res := db.Model(&entity.Link{}).
		Where("url = ? AND status = ? AND count < `limit`", url, true).
		UpdateColumn("count", gorm.Expr("count + 1"))
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}
