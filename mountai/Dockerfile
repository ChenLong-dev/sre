#build stage
#FROM crpi-592g7buyguepbrqd-vpc.cn-shanghai.personal.cr.aliyuncs.com/shanhaii/golang:1.20 AS builder
FROM golang:1.20 AS builder
RUN pwd
WORKDIR /go/src/build
ENV GO111MODULE=on
#ENV GOPROXY=http://goproxy.shanhai.int:8081
ENV GOPROXY=https://goproxy.cn,direct
#ENV GONOSUMDB==gitlab.shanhai.int
ENV GOSUMDB=sum.golang.org
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# =====================
# 根据情况修改主程序包位置
# =====================
# 主程序
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o target/http cmd/http/http.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o target/status_worker cmd/status_worker/status_worker.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o target/app_op_consumer cmd/app_op_consumer/app_op_consumer.go
# RUN for file in cmd/*/*.go; do fileName=${file##*/}; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o target/${fileName%.*} $file; done
# 脚本
# RUN for file in scripts/*/*.go; do fileName=${file##*/}; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o target/${fileName%.*} $file; done

#run stage
#FROM crpi-592g7buyguepbrqd-vpc.cn-shanghai.personal.cr.aliyuncs.com/shanhaii/alpine:3.10
FROM alpine:3.11.5
RUN pwd
WORKDIR /root
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add --no-cache -U ca-certificates
RUN apk add --no-cache -U curl
RUN apk add --no-cache -U graphviz
COPY --from=builder /go/src/build/target/* ./
COPY --from=builder /go/src/build/template ./template
COPY --from=builder /go/src/build/config ./config
