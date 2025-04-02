FROM golang:1.24.1 AS builder

WORKDIR /file_storage_api

COPY go.mod ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /fileStorage

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /fileStorage /fileStorage

EXPOSE 8082

CMD ["/fileStorage"]

