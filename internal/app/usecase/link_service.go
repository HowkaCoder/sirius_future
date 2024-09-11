package usecase

import (
	"encoding/json"
	"fmt"
	"sirius_future/internal/app/entity"
	"sirius_future/internal/app/repository"
	"sirius_future/internal/app/service"
	"time"
)

type FutureSiriusUsecase interface {
	CreateLink(userID uint, limit uint) (string, error)
	CheckTheLink(url string) (bool, error)
	CreateUser(user *entity.User) error

	GetReferrerByUrl(url string) (*entity.User, error)
	GetAllLinks() ([]entity.Link, error)
	GetAllUsers() ([]entity.User, error)

	CreatePayment(payment *entity.Payment) error
	GetAllPayments() ([]entity.Payment, error)
	GetPaymentsByUserID(id uint) ([]entity.Payment, error)
	UpdatePayment(id uint, payment *entity.Payment) error
}

type futureSiriusUsecase struct {
	repo    repository.FutureSiriusRepository
	service service.FutureSiriusService
	redis   service.RedisService
}

func NewFutureSiriusUsecase(repo repository.FutureSiriusRepository, service service.FutureSiriusService, redis service.RedisService) *futureSiriusUsecase {
	return &futureSiriusUsecase{repo: repo, service: service, redis: redis}
}
func (fsu *futureSiriusUsecase) GetPaymentsByUserID(userID uint) ([]entity.Payment, error) {
	cacheKey := fmt.Sprintf("user_payments_%d", userID)

	cachedData, err := fsu.redis.Get(cacheKey)
	if err == nil && cachedData != "" {
		var payments []entity.Payment
		err = json.Unmarshal([]byte(cachedData), &payments)
		if err == nil {
			return payments, nil
		}
	}

	payments, err := fsu.repo.GetPaymentsByUserID(userID)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(payments)
	if err == nil {
		fsu.redis.Set(cacheKey, data, 3*time.Hour)
	}

	return payments, nil
}

func (fsu *futureSiriusUsecase) GetAllPayments() ([]entity.Payment, error) {
	data, err := fsu.redis.Get("all_payments")
	if err == nil && data != "" {
		var payments []entity.Payment
		err = json.Unmarshal([]byte(data), &payments)
		if err == nil {
			return payments, nil
		}
	}

	pAyments, err := fsu.repo.GetAllPayments()
	if err != nil {
		return nil, err
	}

	result, err := json.Marshal(pAyments)
	if err == nil {
		fsu.redis.Set("all_payments", result, 3*time.Hour)
	}

	return pAyments, nil

}
func (fsu *futureSiriusUsecase) CreatePayment(payment *entity.Payment) error {
	if err := fsu.repo.CreatePayment(payment); err != nil {
		return err
	}

	payments, _ := fsu.repo.GetAllPayments()

	data, err := json.Marshal(payments)
	if err == nil {
		fsu.redis.Set("all_payments", data, 3*time.Hour)
	}

	return nil
}

func (fsu *futureSiriusUsecase) UpdatePayment(id uint, payment *entity.Payment) error {
	if err := fsu.repo.UpdatePayment(id, payment); err != nil {
		return err
	}

	payments, _ := fsu.repo.GetAllPayments()

	data, err := json.Marshal(payments)
	if err == nil {
		fsu.redis.Set("all_payments", data, 3*time.Hour)
	}

	return nil
}

func (fsu *futureSiriusUsecase) GetAllLinks() ([]entity.Link, error) {
	cachedData, err := fsu.redis.Get("all_links")
	if err == nil && cachedData != "" {
		var links []entity.Link
		err = json.Unmarshal([]byte(cachedData), &links)
		if err == nil {
			return links, nil
		}
	}

	links, err := fsu.repo.GetAllLinks()
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(links)
	if err == nil {
		fsu.redis.Set("all_links", data, 3*time.Hour)
	}

	return links, nil

}

func (fsu *futureSiriusUsecase) GetAllUsers() ([]entity.User, error) {
	cachedData, err := fsu.redis.Get("all_users")
	if err == nil && cachedData != "" {
		var users []entity.User
		err = json.Unmarshal([]byte(cachedData), &users)
		if err == nil {
			return users, nil
		}
	}

	users, err := fsu.repo.GetAllUsers()
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(users)
	if err == nil {
		fsu.redis.Set("all_users", data, 3*time.Hour)
	}

	return users, nil
}
func (fsu *futureSiriusUsecase) GetReferrerByUrl(url string) (*entity.User, error) {
	return fsu.repo.GetReferrerByUrl(url)
}

func (fru *futureSiriusUsecase) CreateLink(userID uint, limit uint) (string, error) {

	if err := fru.service.CheckUserByID(userID); err != nil {
		return "", err
	}

	url := fru.service.GenerateRefferalLink(userID)

	Link := &entity.Link{
		Url:        url,
		ReferrerID: userID,
		Limit:      limit,
	}
	if err := fru.repo.CreateLink(Link); err != nil {
		return "", err
	}

	links, _ := fru.repo.GetAllLinks()

	data, err := json.Marshal(links)
	if err == nil {
		fru.redis.Set("all_links", data, 3*time.Hour)
	}

	return url, nil
}

func (fru *futureSiriusUsecase) CheckTheLink(url string) (bool, error) {
	return fru.repo.CheckTheLink(url)
}

func (fru *futureSiriusUsecase) CreateUser(user *entity.User) error {
	validateErrors := fru.service.UserValidate(user)
	if len(validateErrors) > 0 {
		for _, value := range validateErrors {
			return fmt.Errorf("User Validate Error :%s", value)
		}
	}

	if err := fru.repo.CreateUser(user); err != nil {
		return err
	}

	users, _ := fru.repo.GetAllUsers()

	data, err := json.Marshal(users)
	if err == nil {
		fru.redis.Set("all_users", data, 3*time.Hour)
	}

	return nil
}
