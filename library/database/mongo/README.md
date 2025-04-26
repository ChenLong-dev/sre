# mongo

## 基本用途

1. mongo数据库工具，底层使用 https://github.com/mongodb/mongo-go-driver
2. 具体的配置见Config注释

## 日志渲染模版

使用方式见logrender包

默认渲染模版为 %J{tsTUSM}

以下为当前包支持的格式化字符

* %T：结束时间
* %S：打印日志的调用源
* %s：开始时间
* %U：context中的uuid
* %t：日志标题
* %M：汇总的mongo参数
* %D：数据源名称
* %d：操作持续时间
* %N：db名称
* %n：集合名称
* %F：调用函数名称
* %f：过滤的字段
* %C：改变的字段
* %E：额外的字段，如聚合管道
* %O：参数字段

## 示例

见example_test.go的example