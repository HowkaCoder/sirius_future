package main

import (
	"log"
	"net/http"
	"sirius_future/internal"
	"sirius_future/internal/app/handler"
	"sirius_future/internal/app/repository"
	"sirius_future/internal/app/service"
	"sirius_future/internal/app/usecase"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"

	"github.com/gofiber/fiber/v2"

	"gorm.io/gorm"
)

var DB *gorm.DB

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "method", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method", "status"},
	)
)

func prometeus_init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
}

var jwtSecret = []byte("supersecretkey")

func main() {
	DB = internal.DatabaseInit()
	logger := service.InitLogger()
	redisService := service.NewRedisService("localhost:6379")

	logService := service.NewLoggerService(logger)

	FutureSiriusRepo := repository.NewFutureSiriusRepository(DB, logService)
	FutureSiriusService := service.NewFutureSiriusService(DB)
	FutureSiriusUsecase := usecase.NewFutureSiriusUsecase(FutureSiriusRepo, FutureSiriusService, *redisService)
	FutureSiriusHandler := handler.NewLinkHandler(FutureSiriusUsecase)

	app := fiber.New()

	// Middleware for Prometheus metrics
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		status := c.Response().StatusCode()
		httpRequestsTotal.WithLabelValues(c.Path(), c.Method(), http.StatusText(status)).Inc()
		httpRequestDuration.WithLabelValues(c.Path(), c.Method(), http.StatusText(status)).Observe(time.Since(start).Seconds())

		return err
	})

	// Public routes
	app.Post("/api/create-link", FutureSiriusHandler.CreateLink)
	app.Get("/api/check-link/:url", FutureSiriusHandler.CheckTheLink)
	app.Post("/register", FutureSiriusHandler.CreateUserWithoutLink)
	/*app.Post("/login", func(c *fiber.Ctx) error {
		// Simplified authentication
		type LoginRequest struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.ErrBadRequest
		}

		// Здесь должна быть проверка пользователя в базе данных
		// Для примера используем "admin" как username и password
		if req.Username == "admin" && req.Password == "admin" {
			token, err := generateJWT(req.Username)
			if err != nil {
				return fiber.ErrInternalServerError
			}
			return c.JSON(fiber.Map{"token": token})
		}

		return fiber.ErrUnauthorized
	})
	*/
	// JWT-protected routes

	app.Get("/api/users", FutureSiriusHandler.GetAllUsers)
	app.Get("/api/links", FutureSiriusHandler.GetAllLinks)
	app.Get("/api/get-referrer/:url", FutureSiriusHandler.GetReferrerByUrl)

	app.Post("/api/payments", FutureSiriusHandler.CreatePayment)
	app.Get("/api/payments", FutureSiriusHandler.GetAllPayments)
	app.Get("/api/payments/user/:id", FutureSiriusHandler.GetPaymentsByUserID)
	app.Patch("/api/payments/:id", FutureSiriusHandler.UpdatePayment)

	// Prometheus metrics
	app.Get("/metrics", func(c *fiber.Ctx) error {
		fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())(c.Context())
		return nil
	})

	log.Fatal(app.Listen(":8080"))
}
