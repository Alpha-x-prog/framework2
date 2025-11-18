FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o api_gateway .

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/api_gateway /app/api_gateway

EXPOSE 8080

ENV APP_ENV=dev

CMD ["./api_gateway"]
