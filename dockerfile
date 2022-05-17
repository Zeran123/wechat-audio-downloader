FROM golang:latest
WORKDIR /www

ENV GOPROXY=https://proxy.golang.com.cn,direct

COPY go.mod /www/go.mod
COPY go.sum /www/go.sum
COPY main.go /www/main.go

EXPOSE 8080

RUN go mod tidy

ENTRYPOINT go run .