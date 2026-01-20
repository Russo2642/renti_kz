# Multi-stage Dockerfile для Go приложения renti_kz

# Стадия 1: Сборка приложения
FROM golang:1.23.0-alpine AS builder

# Устанавливаем git и необходимые инструменты
RUN apk add --no-cache git ca-certificates tzdata

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение с оптимизацией
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo -o main cmd/api/main.go

# Стадия 2: Финальный образ
FROM alpine:latest

# Устанавливаем необходимые сертификаты, timezone данные и wget для healthcheck
RUN apk --no-cache add ca-certificates tzdata wget curl

# Создаем пользователя для безопасности
RUN addgroup -g 1001 -S renti && \
    adduser -u 1001 -S renti -G renti

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем бинарный файл из builder стадии
COPY --from=builder /app/main .

# Копируем миграции базы данных
COPY --from=builder /app/migrations ./migrations

# Копируем шаблоны договоров
COPY --from=builder /app/internal/templates ./internal/templates

# Создаем папку для конфигурации
RUN mkdir -p /app/config

# Создаем папку для uploads
RUN mkdir -p /app/uploads

# Меняем владельца файлов сначала
RUN chown -R renti:renti /app

# Устанавливаем права на выполнение
RUN chmod +x main

# Устанавливаем права на директории и файлы шаблонов
RUN chmod -R 755 /app/internal/templates
RUN find /app/internal/templates -type f -exec chmod 644 {} \;

# Переключаемся на непривилегированного пользователя
USER renti

# Открываем порт 8080
EXPOSE 8080

# Определяем health check (используем wget который мы установили)
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/swagger/index.html || exit 1

# Запускаем приложение
CMD ["./main"] 