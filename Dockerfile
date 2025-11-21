# 1) билд
FROM golang:1.24-alpine AS builder

WORKDIR /app

# зависимости
RUN apk add --no-cache git

# копируем для кэширования слоев и подгружаем
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# собираем
RUN go build -o pr-reviewer cmd/app/main.go



# 2) запуск
FROM alpine:3.18

WORKDIR /app

# копируем бинарник с прошлого этапа
COPY --from=builder /app/pr-reviewer .

# миграции
COPY migrations ./migrations

# окружение по умолчанию
ENV PORT=8080
ENV DB_HOST=db
ENV DB_USER=db_pr_user
ENV DB_PASSWORD=pr_secret
ENV DB_NAME=db_pr_name

# порт
EXPOSE 8080

# запускаемся
CMD ["./pr-reviewer"]