FROM golang:1.24.1 AS builder

WORKDIR /notification_api

COPY go.mod ./
RUN go mod download 

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /notification

EXPOSE 8083

CMD ["/notification"]
