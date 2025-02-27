# to build this docker image:
#   docker build -f auth.Dockerfile -t landaiqing/schisandra-auth-server:v1.0.0 .

FROM golang:1.23.5-bullseye AS builder

LABEL maintainer="landaiqing <<landaiqing@126.com>>"

ENV TZ=Asia/Shanghai \
    DEBIAN_FRONTEND=noninteractive

RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /app

COPY . .

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn,direct

RUN go mod download &&  \
    go build -ldflags="-w -s" -o schisandra-auth-server ./app/auth/api/auth.go


FROM alpine:latest

ENV TZ=Asia/Shanghai \
    DEBIAN_FRONTEND=noninteractive

RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone && \
    apk add --no-cache tzdata libjpeg-turbo && \
    mkdir -p /app/api/etc

WORKDIR /app

COPY --from=builder /app/schisandra-auth-server .

COPY --from=builder /app/app/auth/resources ./resources

EXPOSE 80

CMD ["./schisandra-auth-server"]


# To run this docker image:
#   docker run -p 80:80 -v /home/schisandra/backed/auth/etc:/app/api/etc --name schisandra-auth-server --restart=always landaiqing/schisandra-auth-server:v1.0.0