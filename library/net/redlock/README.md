# redlock

## 基本用途

1. 基于redis的分布式锁，底层使用 https://github.com/go-redsync/redsync
2. 具体的配置见Config注释

## 日志渲染模版

使用方式见logrender包

默认渲染模版为 %J{tbTUSR}

以下为当前包支持的格式化字符

* %T：结束时间
* %b：开始时间
* %S：打印日志的调用源
* %U：context中的uuid
* %t：日志标题
* %R：汇总的redlock参数
* %D：操作持续时间
* %s：状态
* %N：锁名称
* %n：操作名称

## 示例

见example_test.go的example