package runtime

import (
	"fmt"
	"regexp"
	"runtime"
	"strconv"
)

var (
	// 过滤调用栈的正则数组
	DefaultFilterCallerRegexp = []*regexp.Regexp{
		LibraryCallerRegexp,
		FrameworkCallerRegexp,
		GormCallerRegexp,
	}
	// gorm的调用栈过滤
	GormCallerRegexp = regexp.MustCompile(`jinzhu/gorm(@.*)?/.*.go`)
	// library的调用栈过滤
	LibraryCallerRegexp = regexp.MustCompile(`gitlab.shanhai.int/sre/library(@.*)?/.*.go`)
	// framework的调用栈过滤
	FrameworkCallerRegexp = regexp.MustCompile(`gitlab.shanhai.int/sre/app-framework(@.*)?/.*.go`)
)

// 获取调用栈中的单行
func GetCaller(skip int) (name string) {
	if _, file, lineNo, ok := runtime.Caller(skip); ok {
		return file + ":" + strconv.Itoa(lineNo)
	}
	return "unknown:0"
}

// 获取完整调用栈
func GetFullCallers() []string {
	var pcs [32]uintptr
	n := runtime.Callers(0, pcs[:])

	callers := make([]string, 0)
	for _, pc := range pcs[0:n] {
		file := "unknown"
		lineNo := 0
		fn := runtime.FuncForPC(pc - 1)
		if fn != nil {
			file, lineNo = fn.FileLine(pc - 1)
		}

		callers = append(callers, fmt.Sprintf("%s:%d ", file, lineNo))
	}

	return callers
}

// 获取过滤后的调用栈
func GetDefaultFilterCallers() string {
	return GetFilterCallers(DefaultFilterCallerRegexp)
}

// 获取过滤后的调用栈
func GetFilterCallers(filter []*regexp.Regexp) string {
	for i := 1; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)

		isFiltered := false
		for _, re := range filter {
			if re.MatchString(file) {
				isFiltered = true
				break
			}
		}
		if ok && !isFiltered {
			return fmt.Sprintf("%v:%v", file, line)
		}
	}
	return ""
}
