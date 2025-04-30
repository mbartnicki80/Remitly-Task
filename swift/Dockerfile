FROM golang:1.24.2 AS builder
WORKDIR /app
COPY . .
RUN GOARCH=amd64 GOOS=linux go build -o main ./cmd/main.go
RUN chmod +x main

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
CMD ["./main"]
EXPOSE 8080
