# kafka

## 基本用途

1. kafka消息队列相关工具，底层使用 https://github.com/yahoo/kafka-manager
2. 具体的配置见Config注释

## 日志渲染模版

使用方式见logrender包

默认渲染模版为 %J{tTK}

以下为当前包支持的格式化字符

* %T：当前时间
* %t：日志标题
* %K：汇总的kafka参数
* %M：日志信息

## 示例

见example_test.go的example