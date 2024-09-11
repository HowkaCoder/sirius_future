package service

import (
	"log"

	"os"

	"golang.org/x/exp/slog"
)

type LoggerService interface {
	Info(msg string, args ...any)
	Error(msg string, err error, args ...any)
	Debug(msg string, args ...any)
}

type loggerService struct {
	log *slog.Logger
}

func NewLoggerService(log *slog.Logger) *loggerService {
	return &loggerService{log: log}
}

func InitLogger() *slog.Logger {

	logFile, err := os.OpenFile("log/sirius_future.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	Logger := slog.New(slog.NewJSONHandler(logFile, nil))
	return Logger
}

func (lg *loggerService) Info(msg string, args ...any) {
	lg.log.Info(msg, args...)
}

func (lg *loggerService) Error(msg string, err error, args ...any) {
	lg.log.Error(msg, append(args, "error", err)...)
}

func (lg *loggerService) Debug(msg string, args ...any) {
	lg.log.Debug(msg, args...)
}
