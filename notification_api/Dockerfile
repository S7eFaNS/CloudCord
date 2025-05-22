FROM golang:1.24.1 AS builder

WORKDIR /notification_api

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /notification && chmod +x /notification

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /notification /notification

EXPOSE 8083
EXPOSE 2112

CMD ["/notification"]

