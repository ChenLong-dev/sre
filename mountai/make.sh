#!/bin/sh
#@auth cl
#@time 20240920

# 获取运行服务器架构名称
ARCH=$(uname -m)

#echo -e "\033[30m ### 30:黑   ### \033[0m"
#echo -e "\033[31m ### 31:红   ### \033[0m"
#echo -e "\033[32m ### 32:绿   ### \033[0m"
#echo -e "\033[33m ### 33:黄   ### \033[0m"
#echo -e "\033[34m ### 34:蓝色 ### \033[0m"
#echo -e "\033[35m ### 35:紫色 ### \033[0m"
#echo -e "\033[36m ### 36:深绿 ### \033[0m"
#echo -e "\033[37m ### 37:白色 ### \033[0m"

# 获取shell脚本运行路径
SHELL_BASE_PATH=$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd)

CGO_ENABLED=0
GOOS=$(go env GOOS)
GOARCH=${ARCH}

function make() {
  echo -e "\033[34m ######################### [make] ${CGO_ENABLED} ${GOOS} ${GOARCH} ########## \033[0m"
  rm -rf target
#  CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${GOARCH} go build -v -o target/http cmd/http/http.go
#  CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${GOARCH} go build -v -o target/status_worker cmd/status_worker/status_worker.go
#  CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${GOARCH} go build -v -o target/app_op_consumer cmd/app_op_consumer/app_op_consumer.go

  go build -v -o target/http cmd/http/http.go
  go build -v -o target/status_worker cmd/status_worker/status_worker.go
  go build -v -o target/app_op_consumer cmd/app_op_consumer/app_op_consumer.go
}




function print_help() {
    echo -e "\033[35m ######################### HELP ARCH:${ARCH} ######################### \033[0m"
    echo -e "\033[35m #sh scp_file.sh {param} \033[0m"
    echo -e "\033[35m {param}: \033[0m"
    echo -e "\033[35m        -m       : make \033[0m"
    echo -e "\033[35m        -r       : run  \033[0m"
    echo -e "\033[35m        -c       : clean \033[0m"
    echo -e "\033[35m        -p       : package \033[0m"
    echo -e "\033[35m        -        : build -> package \033[0m"
    echo -e "\033[35m        - help   : help \033[0m"
    echo -e "\033[35m ######################### HELP ARCH:${ARCH} ######################### \033[0m"
    exit 1
}

function main() {
    echo -e "\033[34m ######################### make.sh input param is $@ ######################### \033[0m"
    case $1 in
        "-m")
            make
            ;;
        "-r")
#            run
            ;;
        "-c")
#            clean
            execute $1
            ;;
        "-p")
            execute $1
            ;;
        "-")
            execute -b
            execute -p
            ;;
        *)
          print_help
          ;;
    esac
}

main "$@"