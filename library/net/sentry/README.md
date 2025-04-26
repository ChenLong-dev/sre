# Sentry

## 基本用途

1. sentry异常处理工具，底层使用 github.com/getsentry/sentry-go
2. 具体的配置见Config注释

## 日志渲染模版

使用方式见logrender包

默认渲染模版为 %J{tLSs}

以下为当前包支持的格式化字符

* %L：当前时间
* %S：打印日志的调用源
* %t：日志标题
* %D：sentry DSN
* %T：标签
* %e：事件ID
* %E：开发环境
* %s：参数汇总

## 示例

见example_test.go的example