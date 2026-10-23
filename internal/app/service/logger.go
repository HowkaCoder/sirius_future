package service

import (
	"log/slog"
	"os"
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
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
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
