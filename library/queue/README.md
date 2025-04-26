# queue

## 基本用途

1. amqp消息队列相关工具，底层使用 https://github.com/streadway/amqp
2. 具体的配置见Config注释

## 日志渲染模版

使用方式见logrender包

默认渲染模版为 %J{tTUA}

以下为当前包支持的格式化字符

* %T：当前时间
* %U：context中的uuid
* %t：日志标题
* %A：汇总的amqp参数
* %N：队列名称
* %x：交换机名
* %r：路由key
* %C：消息类型
* %B：消息体
* %e：额外消息
* %E：错误信息

## 示例

见example_test.go的example