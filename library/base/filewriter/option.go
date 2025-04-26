package filewriter

import (
	"fmt"
	"strings"
	"time"
)

// RotateFormat
const (
	RotateDaily = "2006-01-02"
)

// 默认配置
var defaultOption = option{
	// 默认分割文件的时间格式化模版
	RotateFormat: RotateDaily,
	// 默认1G
	MaxSize: 1 << 30,
	// 默认64k
	ChanSize: 1024 * 8,
	// 默认10s
	RotateInterval: 10 * time.Second,
}

// 配置
type option struct {
	// 分割文件的时间格式化模版
	RotateFormat string
	// 最大文件数量
	MaxFile int
	// 单个文件最大大小
	MaxSize int64
	// 缓冲区大小
	ChanSize int
	// 分割检查时间间隔
	RotateInterval time.Duration
	// 写入超时时间
	WriteTimeout time.Duration
}

// 配置的应用函数
type OptionFunc func(opt *option)

// RotateFormat e.g 2006-01-02 meaning rotate log file every day.
// NOTE: format can't contain ".", "." will cause panic ヽ(*。>Д<)o゜.
func RotateFormat(format string) OptionFunc {
	if strings.Contains(format, ".") {
		panic(fmt.Sprintf("rotate format can't contain '.' format: %s", format))
	}
	return func(opt *option) {
		opt.RotateFormat = format
	}
}

// 最大文件数量
func MaxFile(n int) OptionFunc {
	return func(opt *option) {
		opt.MaxFile = n
	}
}

// 单个文件最大大小
func MaxSize(n int64) OptionFunc {
	return func(opt *option) {
		opt.MaxSize = n
	}
}

// 缓冲区大小，用于内部对象池等
// 如设置过大会导致filewriter占用较多内存直至程序退出
func ChanSize(n int) OptionFunc {
	return func(opt *option) {
		opt.ChanSize = n
	}
}
