# httpclient

## 基本用途

1. http客户端
2. 具体的配置见Config注释

## 日志渲染模版

使用方式见logrender包

默认渲染模版为 %J{taTUSH}

以下为当前包支持的格式化字符

* %T：结束时间
* %a：开始时间
* %S：打印日志的调用源
* %U：context中的uuid
* %t：日志标题
* %H：汇总的http client参数
* %s：状态码
* %B：响应body
* %D：请求时间
* %u：请求url
* %e：请求endpoint
* %h：请求头
* %b：请求body
* %M：请求方法
* %E：额外信息

## 示例

见example_test.go的example