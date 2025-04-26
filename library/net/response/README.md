# response

## 基本用途

1. 用于response响应
2. 支持自定义响应
3. 服务中所有响应必须使用该包
4. 旧服务响应建议在服务内定义response包，并使用自定义响应，详情见example
5. 若服务产生错误，建议http请求码返回200，通过response中的错误码和消息去区分

## 示例

见example_test.go的example