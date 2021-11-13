# syntax=docker/dockerfile:1
FROM golang:1.16-alpine
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


# 暴露 8443 用于公共 Bot 的 Oauth 与回调，不需要可以注释
# Export for PublicNotinBot's Oauth And Callback, you can annotation it if you not need it.
EXPOSE 8443

CMD  [ "/docker-martrix-channel"]
