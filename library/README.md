# Golang-library 基础依赖库

## 目录

以下为基础目录，每个包的使用方法见包下ReadMe说明及Example
- base:基础组件
    - ctime:自定义时间类型
    - decimal:小数处理常用工具
    - deepcopy:深拷贝结构体v1，已废弃
    - deepcopy_v2:深拷贝结构体v2
    - encrypt:加密常用工具
    - filewriter:可分割文件的writer
    - hook:组件统一注入工具
    - net:网络常用工具
    - logrender:日志输出渲染工具
    - null: 空值处理工具
    - reflect:反射常用工具
    - runtime:运行时常用工具
    - slice:切片处理常用工具
    - strings:字符串处理常用工具
    - sw:滑动窗口算法工具
- database:数据库组件
    - etcd:etcd客户端
    - mongo:mongo客户端
    - redis:redis客户端
    - sql:mysql客户端
- goroutine:协程组件
- kafka:Kafka消息队列组件
- log:日志组件
- net:网络组件
    - agollo:apollo组件
    - circuitbreaker:断路器
    - cm:配置中心客户端
    - errcode:错误码
    - gin:gin常用工具
    - httpclient:http客户端
    - metric:数据监控工具
    - middleware:中间件
    - redlock:redis分布式锁工具
    - request:基础请求
    - response:基础响应
    - sentry:sentry异常捕获组件
    - tracing:链路跟踪工具
    - trafficshaping:限流工具
- queue:AMQP消息队列组件