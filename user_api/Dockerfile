FROM golang:1.24.1 AS builder

WORKDIR /user_api

COPY go.mod ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /user

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /user /user

EXPOSE 8081

CMD ["/user"]

