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

RUN apk add --no-cache postgresql-client bash curl
RUN apk add --no-cache curl bash && \
    ARCH=$(uname -m) && \
    case $ARCH in \
        x86_64) MIGRATE_ARCH="linux-amd64";; \
        aarch64) MIGRATE_ARCH="linux-arm64";; \
        *) echo "unsupported architecture: $ARCH" && exit 1;; \
    esac && \
    curl -L "https://github.com/golang-migrate/migrate/releases/download/v4.16.0/migrate.$MIGRATE_ARCH.tar.gz" \
    -o migrate.tar.gz && \
    tar -xzf migrate.tar.gz -C /tmp && \
    mv /tmp/migrate /usr/local/bin/migrate && \
    rm migrate.tar.gz

# копируем бинарник с прошлого этапа
COPY --from=builder /app/pr-reviewer .

# миграции
COPY migrations ./migrations

# порт
EXPOSE 8080

# запускаемся
CMD ["./pr-reviewer"]