FROM golang:1.25-alpine AS builder

WORKDIR /app

# Зависимости
COPY go.mod go.sum ./
RUN go mod download

# Код
COPY . .

# Сборка
RUN CGO_ENABLED=0 GOOS=linux go build -o post-service ./cmd/server

# --- Финальный образ ---
FROM alpine:3.19

WORKDIR /app

# Бинарник
COPY --from=builder /app/post-service .

# Миграции
COPY --from=builder /app/migration/migrations ./migrations

EXPOSE 8083

CMD ["./post-service"]
