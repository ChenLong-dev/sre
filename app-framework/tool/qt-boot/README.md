# qt-boot app-framework服务基础框架脚手架

## 基础说明
1. 脚手架工具，用于快速生成可修改的基础服务器代码
2. 可以直接编译运行

## 使用说明

### 安装
1. 环境变量设置
   ```
   GOPROXY=http://goproxy.qingting-hz.com
   
   GOSUMDB=off
   ```
2. 通过go get安装
   ```
   go get -u gitlab.shanhai.int/sre/app-framework/tool/qt-boot
   ```
   
### 运行
1. 生成名为“testapp”的应用
   ```
   qt-boot new testapp
   ```
2. 进入testapp文件夹，运行“go run cmd/http/http.go”，访问网址“http://localhost:8080/v1/api”验证基础代码运行正常

3. 获取帮助信息
   ```
   qt -h 获取子命令信息
   qt cmd -h 获取子命令cmd的可选项信息
   ```
4. 获取项目组件信息，阅读app-framework/internal/example中示例代码与README
