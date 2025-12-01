# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY . .
RUN go mod tidy
RUN go build -o main ./cmd/main.go

# Run stage
FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/docs ./docs

EXPOSE 8000

CMD ["./main"]
