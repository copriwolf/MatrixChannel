# syntax=docker/dockerfile:1
FROM golang:1.16
# 为我们的镜像设置必要的环境变量
# ENV GO111MODULE=on \
#      GOPROXY=https://goproxy.cn \
#     CGO_ENABLED=0 \
#     GOOS=linux \
#     GOARCH=amd64

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY ./ .
RUN go build -o /docker-martrix-channel
CMD  [ "/docker-martrix-channel"]
