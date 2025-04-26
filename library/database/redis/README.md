# redis

## 基本用途

1. redis数据库工具，底层使用 https://github.com/gomodule/redigo
2. 具体的配置见Config注释

## 日志渲染模版

使用方式见logrender包

默认渲染模版为 %J{tsTUSR}

以下为当前包支持的格式化字符

* %T：结束时间
* %S：打印日志的调用源
* %s：开始时间
* %U：context中的uuid
* %t：日志标题
* %R：汇总的redis参数
* %D：操作持续时间
* %N：调用函数名称
* %n：调用命令名称
* %a：调用命令参数
* %r：操作响应

## 示例

见example_test.go的example