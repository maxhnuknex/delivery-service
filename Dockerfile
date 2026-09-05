FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o delivery-service ./cmd/app/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/delivery-service .
EXPOSE 8080
CMD ["./delivery-service"]
