FROM golang:1.24.1 AS builder

WORKDIR /chat_api

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /chat && chmod +x /chat

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /chat /chat


EXPOSE 8084
EXPOSE 2112

CMD ["/chat"]

