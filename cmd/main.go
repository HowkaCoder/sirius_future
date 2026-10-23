package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sirius_future/internal"
	"sirius_future/internal/app/handler"
	"sirius_future/internal/app/repository"
	"sirius_future/internal/app/service"
	"sirius_future/internal/app/usecase"
	"strconv"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"gorm.io/gorm"
)

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

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)
}

type config struct {
	Port       string
	DBPath     string
	RedisAddr  string
	AdminToken string
}

func loadConfig() config {
	cfg := config{
		Port:       os.Getenv("PORT"),
		DBPath:     os.Getenv("DB_PATH"),
		RedisAddr:  os.Getenv("REDIS_ADDR"),
		AdminToken: os.Getenv("ADMIN_TOKEN"),
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.DBPath == "" {
		cfg.DBPath = "database/sirius.db"
	}
	return cfg
}

func main() {
	cfg := loadConfig()
	if cfg.AdminToken == "" {
		log.Println("ADMIN_TOKEN is not set, admin endpoints are disabled")
	}

	db, err := internal.DatabaseInit(cfg.DBPath)
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	app := newApp(db, cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		app.ShutdownWithTimeout(10 * time.Second)
	}()

	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}

func newApp(db *gorm.DB, cfg config) *fiber.App {
	logService := service.NewLoggerService(service.InitLogger())
	repo := repository.NewFutureSiriusRepository(db, logService)
	uc := usecase.NewFutureSiriusUsecase(repo, service.NewFutureSiriusService(db), *service.NewRedisService(cfg.RedisAddr))
	h := handler.NewLinkHandler(uc)

	app := fiber.New()
	app.Use(recover.New())
	app.Use(metrics)

	app.Get("/metrics", func(c *fiber.Ctx) error {
		fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())(c.Context())
		return nil
	})

	app.Post("/register", h.Register)
	app.Get("/api/check-link/:url", h.CheckTheLink)
	app.Get("/api/get-referrer/:url", h.GetReferrerByUrl)

	admin := app.Group("/api", handler.AdminOnly(cfg.AdminToken))
	admin.Post("/create-link", h.CreateLink)
	admin.Get("/users", h.GetAllUsers)
	admin.Get("/links", h.GetAllLinks)
	admin.Post("/payments", h.CreatePayment)
	admin.Get("/payments", h.GetAllPayments)
	admin.Get("/payments/user/:id", h.GetPaymentsByUserID)
	admin.Patch("/payments/:id", h.UpdatePayment)

	return app
}

func metrics(c *fiber.Ctx) error {
	start := time.Now()
	err := c.Next()

	status := strconv.Itoa(c.Response().StatusCode())
	path := c.Route().Path
	httpRequestsTotal.WithLabelValues(path, c.Method(), status).Inc()
	httpRequestDuration.WithLabelValues(path, c.Method(), status).Observe(time.Since(start).Seconds())
	return err
}
