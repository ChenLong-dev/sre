# errcode

## 基本用途

1. 异常相关工具
2. 服务中所有异常错误码必须使用该包
3. 基础错误码必须唯一，返回到前端的错误码可重复

## 常规业务开发流程

1. 首先需要申请相应的错误码段
2. 开发过程中，可以先在业务代码内定义需要的错误码并引用

## 错误码段及公共错误码

请到[Confluence](https://cfc.qingtingfm.com/pages/viewpage.action?pageId=92721937)上查看

## 示例

见example_test.go的example