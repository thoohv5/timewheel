FROM golang:alpine as builder

WORKDIR /go/src/server
COPY . .

RUN echo "https://mirrors.aliyun.com/alpine/v3.6/main/" > /etc/apk/repositories && apk add gcc g++ make libffi-dev openssl-dev libtool git
#RUN export GO111MODULE=on && export GOPROXY=https://goproxy.cn,direct && mkdir -p bin/ && go build -o bin/ ./...
RUN make setup && make build

FROM alpine:latest
LABEL MAINTAINER="lirui@thooh.com"

WORKDIR /go/src/server

COPY --from=builder /go/src/server/bin ./

EXPOSE 8000
VOLUME /data/conf

CMD ["./server", "-conf", "/data/conf"]
