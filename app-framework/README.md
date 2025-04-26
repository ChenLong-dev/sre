# Golang-APP-Framework 服务基础框架

## 说明:

1. Golang 版本要求最低 1.13
2. 使用该库需要设置如下环境变量
   ```
   GOPROXY=http://goproxy.shanhai.int:8081
   
   GOSUMDB=off
   ```
3. 可通过如下指令引入框架
   ```
   go get gitlab.shanhai.int/sre/app-framework@vx.x.x
   ```
4. 使用脚手架工具qt-boot生成项目framework-example的基础代码
   ```
   go get -u gitlab.shanhai.int/sre/app-framework/tool/qt-boot
   qt-boot new framework-example
   ```
   更多qt-boot功能见/tool/qt-boot/README
   更多服务组件信息见/internal/exmaple下示例代码与README
5. 本框架内部基于Library库，如有具体疑问，请查看[Library文档](https://gitlab.shanhai.int/sre/library)

## 示例

见internal包下example，其中包含该框架的基础使用方式及大量常规场景下的使用标准

## 版本说明

[版本说明](https://gitlab.shanhai.int/sre/app-framework/releases)