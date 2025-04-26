# deepcopy

## 基本用途

1. 用于从源结构体深拷贝至目标结构体
2. 常用于view层与model层之间的模型转换
3. 支持多种标签，每种标签的用法见具体示例
4. 内部使用反射，会略微影响性能，如业务接口对性能十分敏感，请手动拷贝

## 标签

标签以deepcopy作为关键词，标签值以':'为分隔符，不同标签间以';'为分隔符。

如 
```
Field string  `deepcopy:"from:TestSrcFrom;timeformat:2006/01/02 15:04:05"`
```

详细标签见`tag.go`文件

## 示例

见example_test.go的example