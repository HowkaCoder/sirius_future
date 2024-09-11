# Dockerfile

# Шаг 1: Устанавливаем Go для сборки
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем все файлы проекта
COPY . .

# Сборка Go-приложения
RUN go build -o /sirius_future ./cmd/main.go

# Шаг 2: Минимизируем финальный образ
FROM alpine:latest

WORKDIR /root/

# Копируем собранное Go-приложение из builder-образа
COPY --from=builder /sirius_future .

# Копируем файл базы данных и логов
COPY ./cmd/database/test.db ./database/test.db
COPY ./cmd/log/sirius_future.log ./log/sirius_future.log

# Открываем порт приложения
EXPOSE 8080

# Запускаем приложение
CMD ["./sirius_future"]
