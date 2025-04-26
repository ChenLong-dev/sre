package strings

import (
	"strings"
)

// 蛇形名称转大驼峰名称
func SnakeNameToBigCamelName(name string) string {
	name = strings.Replace(name, "_", " ", -1)
	name = strings.Title(name)
	return strings.Replace(name, " ", "", -1)
}
