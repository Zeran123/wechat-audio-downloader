FROM ubuntu:22.04
# WORKDIR /www

# ENV GOPROXY=https://proxy.golang.com.cn,direct

# COPY go.mod /www/go.mod
# COPY go.sum /www/go.sum
# COPY main.go /www/main.go

COPY ./audio-downloader /audio-downloader
RUN chmod +x /audio-downloader

EXPOSE 8080
