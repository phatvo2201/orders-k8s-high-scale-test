
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o orders-service ./cmd

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/orders-service .

EXPOSE 8000

CMD ["./orders-service"]
