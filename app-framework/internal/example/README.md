# 示例

## 基础说明

1. 该示例包含该框架的基础使用方式及部分使用标准
2. 在http服务接口示例中，包含大量常规场景下的使用标准
3. Golang 版本要求最低 1.13
4. 本框架内部基于Library库，请查看[Library文档](https://gitlab.shanhai.int/sre/library)
5. 如有其他问题，欢迎提[Issue](https://gitlab.shanhai.int/sre/app-framework/issues/new)

## 环境变量说明

1. 使用该库需要设置如下环境变量
   ```
   GOPROXY=http://goproxy.qingting-hz.com
   
   GOSUMDB=off
   ```

## 目录结构
   ```
   .
   ├── .gitignore                       git忽略文件
   ├── Dockerfile                       Docker构建文件
   ├── cmd                              主程序包
   │   ├── amqp                         AMQP程序包
   │   │   └── amqp.go                  AMQP服务主程序
   │   ├── http                         Http程序包
   │   │   └── http.go                  Http服务主程序
   │   ├── job                          定时任务程序包
   │   │   └── job.go                   定时任务服务主程序
   │   ├── kafka                        Kafka程序包
   │   │   └── kafka.go                 Kafka服务主程序
   │   └── rpc                          RPC程序包
   │       └── rpc.go                   RPC服务主程序
   ├── config                           配置文件目录包
   │   ├── config.go                    基础配置代码
   │   └── config-amqp.yaml             AMQP服务示例配置文件
   │   └── config-http.yaml             Http服务示例配置文件
   │   └── config-job.yaml              定时任务示例配置文件
   │   └── config-kafka.yaml            Kafka服务示例配置文件
   │   └── config-rpc.yaml              RPC服务示例配置文件
   ├── dao                              数据库层包
   │   └── dao.go                       基础数据库代码
   ├── models                           模型包
   │   ├── entity                       实体模型包
   │   ├── req                          请求模型包
   │   └── resp                         响应模型包
   ├── server                           服务器包
   │   ├── http                         Http服务器包
   │   │   ├── handler                  Http服务器接口处理包
   │   │   └── server.go                Http服务器基础代码
   │   ├── job                          任务服务器包
   │   │   └── server.go                RPC服务器基础代码
   │   ├── rpc                          RPC服务器包
   │   │   └── server.go                RPC服务器基础代码
   │   └── subscribe                    订阅服务器包
   │       ├── amqp.go                  AMQP服务器基础代码
   │       └── kafka.go                 Kafka服务器基础代码
   └── service                          服务包
   ```  
  
## 框架本地启动说明

1. 将完整的配置文件`config.yaml`拷贝至当前运行目录的`config`文件夹下
2. 启动服务

【**重要**】注意项目仓库中不应保存任何环境配置文件，包括本地配置。

【**注意**】完整配置文件是指包含密码等私有配置的文件，而远程配置中心的生成文件中并不会显式展示密码。

## 框架使用说明
1. 设置GOPROXY及GOSUMDB环境变量
   ```
   GOPROXY=http://goproxy.qingting-hz.com
   
   GOSUMDB=off
   ```

2. 使用脚手架工具生成基础代码，假设项目名为 framework-example
   ```
   go get -u gitlab.shanhai.int/sre/app-framework/tool/qt-boot
   qt-boot new framework-example
   ```

3. config包
    1. config文件夹下包含config.go与config-http.yaml文件
    2. 【**重要**】根据所需组件，选择性更改或删除`config.go`文件下的配置文件。注意项目仓库中不应保存任何配置文件。
    3. 【**重要**】配置文件需要在[配置中心](https://gitlab.shanhai.int/sre/config-center)提交基础配置文件，并通过MergeRequest合并至master使用，详情请参考配置中心ReadMe。
    4. 【**注意**】配置文件可参考当前项目的`/config/config-*.yaml`以及远程配置中心。

4. dao包   
    1. dao文件夹下包含dao.go与测试文件dao_test.go
    2. 根据所需组件，选择性更改或删除`dao.go`文件下的组件
    
5. models包   
    1. models文件夹下包含空的entity,req与resp包，根据需求自行扩展
    
6. service包
    1. service文件夹下包含service.go, 测试文件service_test.go和服务组件示例文件hello_service.go
    2. 根据所需组件，选择性更改或删除`service.go`文件下的组件 
    
7. server包
    1. server文件夹下包含server.go与示例句柄/hander/hello_handler.go
    2. 根据实际情况，在基础代码中，修改路由及中间件等，并添加handler方法，选更改`server.go`文件下的组件
       
8. 主程序
    1. cmd文件夹下包含主程序入口cmd/http/http.go
    2. 根据所需服务，参考(`amqp.go`/`http.go`/`job.go`/`kafka.go`/`rpc.go`)修改替换cmd/http/http.go文件
       
9. 其他

    1. 基础`.gitignore`,`Dockerfile` 文件已存在，根据实际文件结构修改`Dockerfile`文件

## 提示

1. 为了避免自动引入错误的包，建议关闭自动引用
   ```
   例如：
   使用GoLand时，请关闭 `add unambiguous imports on the fly`
   ```
2. 如项目包含多个启动程序，可通过修改 `Tars`环境管理中的执行命令 / `AMS`发布任务中的启动命令 配置。
   ```
   如该示例中包含5个启动程序，若需要启动`http`程序，则可将命令修改为`./http`，如需要启动`job`程序，则修改为`./job`
   ```