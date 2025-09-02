# Многоэтапная сборка для оптимизации размера образа

# Этап 1: Сборка приложения
FROM node:18-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
COPY package*.json ./

# Устанавливаем зависимости (включая dev для сборки)
RUN npm ci --silent

# Копируем исходный код
COPY . .

# Собираем приложение для продакшена
RUN npm run build

# Этап 2: Создание финального образа для статических файлов
FROM alpine:latest AS production

# Создаем директорию для статических файлов
RUN mkdir -p /app/dist

# Копируем собранное приложение из предыдущего этапа
COPY --from=builder /app/dist /app/dist

# Создаем пользователя для безопасности
RUN addgroup -g 1001 -S appgroup && \
    adduser -S appuser -u 1001 -G appgroup

# Устанавливаем права на файлы
RUN chown -R appuser:appgroup /app/dist

# Переключаемся на пользователя
USER appuser

# Рабочая директория
WORKDIR /app/dist

# Команда для сохранения контейнера активным (файлы будут доступны через volume)
CMD ["tail", "-f", "/dev/null"] 