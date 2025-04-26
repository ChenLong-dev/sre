# apollo

## 基本用途

1. Apollo配置watch工具，底层使用 github.com/shima-park/agollo
2. 具体的配置见Config注释

## 日志渲染模版

使用方式见logrender包

默认渲染模版为 %J{tTSP}

以下为当前包支持的格式化字符

* %T：当前时间
* %S：打印日志的调用源
* %t：日志标题
* %E：额外信息
* %A：apollo AppID
* %C：apollo cluster
* %N：监听namespace
* %I：apollo IP
* %c: 配置变化
* %P: apollo事件参数汇总
* %W：watcher类型

## 示例

见example_test.go的example