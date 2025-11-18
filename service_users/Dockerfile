# Стейдж сборки
FROM golang:1.24-alpine AS builder

WORKDIR /app

# модули
COPY go.mod go.sum ./
RUN go mod download

# исходники
COPY . .

# собираем бинарник
RUN go build -o service_users .

# Стейдж рантайма
FROM alpine:3.20

WORKDIR /app

# для sqlite
RUN apk add --no-cache ca-certificates sqlite-libs

COPY --from=builder /app/service_users /app/service_users

# COPY .env.prod .env

EXPOSE 8081

ENV APP_ENV=dev

CMD ["./service_users"]
