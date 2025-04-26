#build stage
FROM golang:1.13-alpine AS builder
WORKDIR /go/src/build
ENV GO111MODULE=on
ENV GOPROXY=http://goproxy.qingting-hz.com
ENV GOSUMDB=off
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# =====================
# 根据情况修改主程序包位置
# =====================
RUN for file in internal/example/cmd/*/*.go; do fileName=${file##*/}; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o target/${fileName%.*} $file; done

# =====================
# gRPC 健康检查工具，其他服务可以忽略
# =====================
RUN go get github.com/grpc-ecosystem/grpc-health-probe

#run stage
FROM alpine:3.10
WORKDIR /root
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add --no-cache -U ca-certificates
RUN apk add --no-cache -U curl
RUN apk add --no-cache -U graphviz
COPY --from=builder /go/src/build/target/* ./

# =====================
# gRPC 健康检查工具，其他服务可以忽略
# =====================
ARG GRPC_HEALTH_PROBE_FILE=grpc-health-probe
COPY --from=builder /go/bin/${GRPC_HEALTH_PROBE_FILE} /bin/${GRPC_HEALTH_PROBE_FILE}
RUN chmod +x /bin/${GRPC_HEALTH_PROBE_PATH}
