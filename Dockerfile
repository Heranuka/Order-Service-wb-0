# Stage 1: Сборка
FROM golang:1.24.5-alpine AS builder

# Обновляем систему и устанавливаем зависимости для сборки
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Копируем go.mod и go.sum для кэширования модулей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем бинарь для Linux AMD64 (убирается динамическая линковка)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o main ./cmd/app/main.go

# Stage 2: Минималистичный runtime образ
FROM alpine:latest

# Опционально добавляем libc для совместимости, если нужно
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Копируем бинарь из стадии сборки
COPY --from=builder /app/main .
COPY .env .
# Копируем конфигурационный файл (или можно смонтировать внешним томом)
COPY local.yaml .

EXPOSE 8080

# Устанавливаем entrypoint и команду
ENTRYPOINT ["./main"]
CMD ["-config", "/app/local.yaml"]
