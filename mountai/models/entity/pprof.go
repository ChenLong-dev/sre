package entity

import (
	"flag"
)

// 类型
type PProfType string

const (
	PProfTypeAlloc        PProfType = "allocs"
	PProfTypeBlock        PProfType = "block"
	PProfTypeGoroutine    PProfType = "goroutine"
	PProfTypeHeap         PProfType = "heap"
	PProfTypeMutex        PProfType = "mutex"
	PProfTypeThreadCreate PProfType = "threadcreate"
	PProfTypeProfile      PProfType = "profile"
	PProfTypeTrace        PProfType = "trace"
)

// 行为
type PProfAction string

const (
	PProfActionSvg      PProfAction = "svg"
	PProfActionTree     PProfAction = "tree"
	PProfActionTop      PProfAction = "top"
	PProfActionDownload PProfAction = "download"
)

// pprof用来分析的flag
type PProfAnalyzedFlags struct {
	// 实际flag set
	FlagSet *flag.FlagSet
	// 命令list
	CommandList []string
}

func (f *PProfAnalyzedFlags) Bool(o string, d bool, c string) *bool {
	return f.FlagSet.Bool(o, d, c)
}

func (f *PProfAnalyzedFlags) Int(o string, d int, c string) *int {
	return f.FlagSet.Int(o, d, c)
}

func (f *PProfAnalyzedFlags) Float64(o string, d float64, c string) *float64 {
	return f.FlagSet.Float64(o, d, c)
}

func (f *PProfAnalyzedFlags) String(o, d, c string) *string {
	return f.FlagSet.String(o, d, c)
}

func (f *PProfAnalyzedFlags) StringList(o, d, c string) *[]*string {
	return &[]*string{f.FlagSet.String(o, d, c)}
}

func (f *PProfAnalyzedFlags) ExtraUsage() string {
	return ""
}

func (f *PProfAnalyzedFlags) AddExtraUsage(eu string) {
}

func (f *PProfAnalyzedFlags) Parse(usage func()) []string {
	return f.CommandList
}

// pprof用来分析的ui
type PProfAnalyzedUI struct {
	// 当前命令索引
	CommandIndex int
	// 命令列表
	CommandList []string
}

func (r *PProfAnalyzedUI) ReadLine(prompt string) (string, error) {
	if r.CommandIndex >= len(r.CommandList) {
		return "", nil
	}
	defer func() {
		r.CommandIndex++
	}()

	return r.CommandList[r.CommandIndex], nil
}

func (r *PProfAnalyzedUI) Print(args ...interface{}) {
}

func (r *PProfAnalyzedUI) PrintErr(args ...interface{}) {
}

func (r *PProfAnalyzedUI) IsTerminal() bool {
	return true
}

func (r *PProfAnalyzedUI) WantBrowser() bool {
	return false
}

func (r *PProfAnalyzedUI) SetAutoComplete(complete func(string) string) {
}
