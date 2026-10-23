package handler

import (
	"crypto/subtle"
	"errors"
	"sirius_future/internal/app/entity"
	"sirius_future/internal/app/repository"
	"sirius_future/internal/app/usecase"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type LinkHandler struct {
	usecase usecase.FutureSiriusUsecase
}

func NewLinkHandler(usecase usecase.FutureSiriusUsecase) *LinkHandler {
	return &LinkHandler{usecase: usecase}
}

type registerRequest struct {
	Firstname  string `json:"first_name"`
	Secondname string `json:"second_name"`
	Lastname   string `json:"last_name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	Phone      string `json:"phone"`
	Url        string `json:"url"`
}

func AdminOnly(token string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		got := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
		if token == "" || subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		return c.Next()
	}
}

func writeError(c *fiber.Ctx, err error) error {
	var ve usecase.ValidationError
	switch {
	case errors.As(err, &ve):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": ve.Msg})
	case errors.Is(err, repository.ErrLinkNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	case errors.Is(err, repository.ErrLinkExhausted):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
}

func (lh *LinkHandler) CreateLink(c *fiber.Ctx) error {
	var request struct {
		ID    uint `json:"user_id"`
		Limit uint `json:"link_limit"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	url, err := lh.usecase.CreateLink(request.ID, request.Limit)
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"link": url})
}

func (lh *LinkHandler) CheckTheLink(c *fiber.Ctx) error {
	ok, err := lh.usecase.CheckTheLink(c.Params("url"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(fiber.Map{"result": ok})
}

func (lh *LinkHandler) Register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	user := entity.User{
		Firstname:  req.Firstname,
		Secondname: req.Secondname,
		Lastname:   req.Lastname,
		Email:      strings.ToLower(strings.TrimSpace(req.Email)),
		Phone:      req.Phone,
	}
	if err := lh.usecase.Register(&user, req.Password, req.Url); err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": user.ID, "referrer_id": user.ReferrerID})
}

func (lh *LinkHandler) GetAllUsers(c *fiber.Ctx) error {
	users, err := lh.usecase.GetAllUsers()
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(users)
}

func (lh *LinkHandler) GetAllLinks(c *fiber.Ctx) error {
	links, err := lh.usecase.GetAllLinks()
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(links)
}

func (lh *LinkHandler) GetReferrerByUrl(c *fiber.Ctx) error {
	referrer, err := lh.usecase.GetReferrerByUrl(c.Params("url"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(referrer)
}

func (lh *LinkHandler) CreatePayment(c *fiber.Ctx) error {
	var payment entity.Payment
	if err := c.BodyParser(&payment); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := lh.usecase.CreatePayment(&payment); err != nil {
		return writeError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": payment.ID})
}

func (lh *LinkHandler) GetAllPayments(c *fiber.Ctx) error {
	payments, err := lh.usecase.GetAllPayments()
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(payments)
}

func (lh *LinkHandler) GetPaymentsByUserID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	payments, err := lh.usecase.GetPaymentsByUserID(uint(id))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(payments)
}

func (lh *LinkHandler) UpdatePayment(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payment id"})
	}

	var payment entity.Payment
	if err := c.BodyParser(&payment); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := lh.usecase.UpdatePayment(uint(id), &payment); err != nil {
		return writeError(c, err)
	}
	return c.JSON(fiber.Map{"result": "payment successfully updated"})
}
