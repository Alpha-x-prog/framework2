FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o service_orders .

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates sqlite-libs

COPY --from=builder /app/service_orders /app/service_orders

EXPOSE 8082

ENV APP_ENV=dev

CMD ["./service_orders"]
