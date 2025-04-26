# log

## 基本用途

1. 日志工具
2. 具体的配置见Config注释
3. 服务中所有日志打印必须使用该包

## 日志渲染模版

使用方式见logrender包

默认渲染模版为 %J{tTLUSm}

以下为当前包支持的格式化字符

* %T：当前时间
* %S：打印日志的调用源
* %U：context中的uuid
* %t：日志标题
* %L：日志级别
* %M：日志信息，text文本形式
* %m：日志信息，json形式

## 自定义日志保留键名

在自定义日志方法中，以下为内部保留键名，不允许使用

'time','level','level_value','source','app_id','uuid'

## 示例

见example_test.go的example