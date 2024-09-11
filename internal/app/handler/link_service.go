package handler

import (
	"fmt"
	"sirius_future/internal/app/entity"
	"sirius_future/internal/app/usecase"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type LinkHandler struct {
	usecase usecase.FutureSiriusUsecase
}

func NewLinkHandler(usecase usecase.FutureSiriusUsecase) *LinkHandler {
	return &LinkHandler{usecase: usecase}
}

func (lh *LinkHandler) CreateLink(c *fiber.Ctx) error {
	var request struct {
		ID    uint `json:"user_id"`
		Limit uint `json:"link_limit"`
	}
	err := c.BodyParser(&request)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error": err.Error(),
		})
	}

	result, err := lh.usecase.CreateLink(request.ID, request.Limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"Link": result,
	})
}

func (lh *LinkHandler) CheckTheLink(c *fiber.Ctx) error {
	url := c.Params("url")

	result, err := lh.usecase.CheckTheLink(url)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"result": result,
	})
}

func (lh *LinkHandler) CreateUserWithoutLink(c *fiber.Ctx) error {
	var user *entity.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error": err.Error(),
		})
	}

	fmt.Println(user)

	if err := lh.usecase.CreateUser(user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"result": "user successfully created",
	})
}

func (lh *LinkHandler) CreateUserWithRefferalLink(c *fiber.Ctx) error {
	var request struct {
		User entity.User `json:"user"`
		Url  string      `json:"url"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error":   err.Error(),
			"message": "error with body",
		})
	}

	result, err := lh.usecase.CheckTheLink(request.Url)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error":   err.Error(),
			"message": "invalid url ",
		})
	}

	if !result {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": "invalid link",
		})
	}
	referrer, err := lh.usecase.GetReferrerByUrl(request.Url)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": err.Error(),
		})
	}

	request.User.ReferrerID = referrer.ID

	if err := lh.usecase.CreateUser(&request.User); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"result": "user successfully created",
	})
}

func (lh *LinkHandler) GetAllUsers(c *fiber.Ctx) error {
	users, err := lh.usecase.GetAllUsers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": err.Error(),
		})
	}

	return c.JSON(users)
}

func (lh *LinkHandler) GetAllLinks(c *fiber.Ctx) error {
	links, err := lh.usecase.GetAllLinks()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": err.Error(),
		})
	}

	return c.JSON(links)
}

func (lh *LinkHandler) GetReferrerByUrl(c *fiber.Ctx) error {
	url := c.Params("url")

	referrer, err := lh.usecase.GetReferrerByUrl(url)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error": err.Error(),
		})
	}

	return c.JSON(referrer)
}

func (lh *LinkHandler) CreatePayment(c *fiber.Ctx) error {
	var payment *entity.Payment
	if err := c.BodyParser(&payment); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error": err.Error(),
		})
	}

	if err := lh.usecase.CreatePayment(payment); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"result": "payment successfully created",
	})
}

func (lh *LinkHandler) GetAllPayments(c *fiber.Ctx) error {
	payments, err := lh.usecase.GetAllPayments()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": err.Error(),
		})
	}

	return c.JSON(payments)
}

func (lh *LinkHandler) GetPaymentsByUserID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error": err.Error(),
		})
	}

	paymets, err := lh.usecase.GetPaymentsByUserID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": err.Error(),
		})
	}

	return c.JSON(paymets)
}

func (lh *LinkHandler) UpdatePayment(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error": err.Error(),
		})
	}

	var payment *entity.Payment
	if err := c.BodyParser(&payment); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error": err.Error(),
		})
	}

	if err := lh.usecase.UpdatePayment(uint(id), payment); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"result": "payment successfully updated",
	})
}
