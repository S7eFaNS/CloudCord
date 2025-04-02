FROM golang:1.24.1 AS builder

WORKDIR /notification_api

COPY go.mod ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /notification

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /notification /notification

EXPOSE 8083

CMD ["/notification"]

