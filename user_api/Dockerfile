FROM golang:1.24.1 AS builder

WORKDIR /user_api

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /user && chmod +x /user

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /user /user

COPY .env .env

EXPOSE 8081

CMD ["/user"]

