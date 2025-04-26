package project

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"github.com/urfave/cli/v2"

	"github.com/gobuffalo/packr/v2"
)

// Project 项目信息
type Project struct {
	Name      string // 项目名称
	Dir       string // 项目父目录
	Type      string // 项目类型
	TmplPath  string // 模板路径
	GoVersion string // Go语言版本
}

// 引导信息
const guideInfo = `Project: {{.Name}}
Type: {{.Type}}
Directory: {{.Dir}}
Launch the Project:
	* go into {{.Dir}}
	* execute the command "go run ./cmd/http/http.go"
	* try to visit http://localhost:80/v1/api to see "Hello, {{.Name}}!"
Learn more infomation in the examples of the app-framework.
`

// flag名称
const (
	flagProjectType string = "type"
	flagProjectDir  string = "dir"
)

// 新项目元数据
var newProject Project

// supporting project type
var supportType []string = []string{"http"}

// GenProjectCmd 生成基础服务器
var GenProjectCmd = &cli.Command{
	Name:    "new",
	Aliases: []string{"n"},
	Usage:   "generate a base project, providing a simple router",
	Flags: []cli.Flag{
		&cli.StringFlag{ // 选择应用类型
			Name:    flagProjectType,
			Aliases: []string{"t"},
			Usage: `specify the type of the new project, including: 
					* http`,
			Value: "http",
		},
		&cli.StringFlag{ // 选择项目父文件夹
			Name:    flagProjectDir,
			Aliases: []string{"d"},
			Usage:   `specify the parent directory of the new project`,
			Value:   "./",
		},
	},
	Action: func(c *cli.Context) error {
		fmt.Fprintf(os.Stdout, "Project being generating\n")

		// 参数检测
		projectDir, err := filepath.Abs(c.String(flagProjectDir))
		if err != nil {
			return err
		}

		newProject.Name = c.Args().Get(0)
		newProject.Type = c.String(flagProjectType)
		newProject.Dir = filepath.Join(projectDir, newProject.Name)

		version := runtime.Version()
		idx := strings.LastIndex(version, ".")
		newProject.GoVersion = strings.TrimPrefix(version[:idx], "go")

		err = CheckProjectProperties()
		if err != nil {
			return err
		}

		// 渲染模板
		fmt.Fprintf(os.Stdout, "\t* templates being rendered\n")
		if err := RenderProject(); err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "\t* project created successfully!\n")

		// 输出指示信息
		RenderToStdout(guideInfo, newProject)

		return nil
	},
}

// RenderProject 渲染填充模板
//go:generate packr2
func RenderProject() error {

	var box *packr.Box
	if newProject.Type == "http" {
		box = packr.New("projectBox", "./templates/http")
	}
	if box == nil {
		return errors.New("No Available Templates")
	}

	// 构建项目目录
	if err := os.MkdirAll(newProject.Dir, 0755); err != nil {
		return err
	}

	for _, fname := range box.List() {

		// 读取模板文件
		fileStr, err := box.FindString(fname)
		if err != nil {
			return err
		}

		// 解析填充模板文件
		var byteBuffer []byte
		if strings.HasSuffix(fname, ".tmpl") {
			byteBuffer, err = ParseTmpl(fileStr)
			fname = strings.TrimSuffix(fname, ".tmpl")
			if err != nil {
				return err
			}
		} else {
			byteBuffer = []byte(fileStr)
		}

		// 生成项目文件
		_, err = WriteBaseFile(fname, byteBuffer)
		if err != nil {
			return err
		}

	}

	return nil
}

// ParseTmpl 解析填充tmpl模板文件
func ParseTmpl(tmplStr string) ([]byte, error) {

	tmpl, err := template.New("projectTmpl").Parse(tmplStr)
	if err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	if err = tmpl.Execute(&buffer, newProject); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// WriteBaseFile 已填充模板写入文件
func WriteBaseFile(fname string, buffer []byte) (string, error) {

	if idx := strings.LastIndex(fname, string(os.PathSeparator)); idx > 0 {
		dir := fname[:idx]
		if err := os.MkdirAll(filepath.Join(newProject.Dir, dir), 0755); err != nil {
			return "", err
		}
	}

	fileName := filepath.Join(newProject.Dir, fname)
	return fileName, ioutil.WriteFile(fileName, buffer, 0644)
}

// RenderToStdout 渲染到标准输出
func RenderToStdout(text string, data interface{}) error {
	tmpl, err := template.New("tmpl").Parse(text)
	if err != nil {
		return err
	}
	err = tmpl.Execute(os.Stdout, data)
	if err != nil {
		return err
	}
	return nil
}

// CheckProjectProperties 检测项目属性
func CheckProjectProperties() error {
	if newProject.Name == "" {
		return errors.New("Invalid Parameter: Empty Name")
	}

	flg := false
	for _, t := range supportType {
		if t == newProject.Type {
			flg = true
			break
		}
	}
	if !flg {
		return errors.New("Invalid Parameter: Unsupported Type")
	}

	if _, err := os.Stat(newProject.Dir); err == nil {
		return errors.New("Invalid Parameter: Project Existed")
	}

	return nil
}
