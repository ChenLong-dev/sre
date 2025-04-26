# sql

## 基本用途

1. mysql数据库工具，底层使用 hhttps://github.com/jinzhu/gorm
2. 具体的配置见Config注释

## 日志渲染模版

使用方式见logrender包

默认渲染模版为 %J{tsTUSG}

以下为当前包支持的格式化字符

* %T：结束时间
* %S：打印日志的调用源
* %s：开始时间
* %U：context中的uuid
* %t：日志标题
* %G：汇总的mysql参数
* %D：数据源名称
* %d：操作持续时间
* %R：操作行数
* %L：gorm日志等级
* %F：完整sql

## 示例

见example_test.go的example