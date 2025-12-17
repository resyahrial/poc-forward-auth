FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o book-service main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/book-service .

EXPOSE 8082

CMD ["./book-service"]
