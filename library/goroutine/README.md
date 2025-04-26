# goroutine

## 基本用途

1. 协程工具
2. 具体的配置见Config注释
3. 服务中所有协程启用必须使用该包

## 日志渲染模版

使用方式见logrender包

默认渲染模版为 %J{tTUSG}

以下为当前包支持的格式化字符

* %T：当前时间
* %S：打印日志的调用源
* %U：context中的uuid
* %t：日志标题
* %G：汇总的goroutine参数
* %s：协程状态
* %N：协程组名
* %m: 协程组模式
* %n：协程名
* %I：协程组id
* %i：协程id
* %E：额外参数，如报错信息

## 示例

见example_test.go的example