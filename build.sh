#!/bin/bash
export GOPROXY=https://proxy.golang.com.cn,direct
go mod tidy
go build -o audio-downloader
