# etcd

## 基本用途

1. etcd数据库工具，底层使用 github.com/coreos/etcd
2. 具体的配置见Config注释

## 日志渲染模版

使用方式见logrender包

默认渲染模版为 %J{tTUSE}

以下为当前包支持的格式化字符

* %T：当前时间
* %S：打印日志的调用源
* %U：context中的uuid
* %t：日志标题
* %E：汇总的etcd参数
* %P：数据前缀
* %K：数据键
* %V：数据值
* %e：额外信息

## 示例

见example_test.go的example