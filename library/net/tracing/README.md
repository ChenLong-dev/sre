# tracing

## 基本用途

1. 链路跟踪相关工具
2. 具体的配置见Config注释

## 日志渲染模版

使用方式见logrender包

默认渲染模版为 %J{tTE}

以下为当前包支持的格式化字符

* %T：当前时间
* %t：日志标题
* %E：汇总的tracing参数
* %L：日志级别
* %M：日志消息

## 示例

见example_test.go的example