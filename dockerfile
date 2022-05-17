FROM golang:latest
WORKDIR /www

COPY go.mod /www/go.mod
COPY main.go /www/main.go

EXPOSE 8080

RUN go run .