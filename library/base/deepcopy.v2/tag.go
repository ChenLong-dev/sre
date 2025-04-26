package deepcopy

import (
	"reflect"
	"strings"
	"sync"
)

const (
	// tag名
	TagName = "deepcopy"
	// tag分割符号
	TagSplitSymbol = ";"
	// tag值分割符号
	TagValueSymbol = ":"
)

// 系统标签
// 默认自动开启
const (
	// 选择器标签

	// 忽略当前字段
	SkipTagName = "skip"
	// 从src结构体的指定方法进行转换
	// 要求方法首个参数为 map[string]interface{} 类型
	// 若最后一个参数为error类型，则会在非空时，抛出
	// 对应的值为方法名
	FieldMethodTagName = "method"
	// 从src结构体的指定字段名称转换
	// 对应的值为src结构体的字段名
	FieldFromTagName = "from"
	// 若源结构体没有符合条件的字段时调用
	// 首先会调用dst结构体的指定方法
	// 若不存在指定方法，则直接赋值给目标字段
	// 目前允许int，int64，uint，uint64，float64，string，bool类型直接赋值
	// 对应的值为默认方法名或默认值
	DefaultMethodName = "default"

	// 转换器标签

	// 从src结构体的time类型格式化为指定string类型
	// 对应的值为time的format格式，例如 `2006/01/02 15:04:05`
	TimeFormatTagName = "timeformat"
	// 从src结构体的任意类型强转为string类型
	// 要求dst结构体字段类型为string
	StringTagName = "string"
	// 从src结构体的int类型转换为bool类型
	// 要求源类型为int，int64，uint，uint64，float64类型
	// 要求dst结构体字段类型为bool
	// 若源值不为0，则转换为true，否则转换为false
	BoolTagName = "bool"
	// primitive.ObjectID 转 string类型
	// 要求dst结构体字段类型string
	// 要求src结构体字段类型primitive.ObjectID
	ObjectIDTagName = "objectid"
)

// 可选标签
// 默认不开启，需手动开启
const (
	// 选择器标签

	// 指定src结构体转换后的目标名称
	// 对应的值为dst结构体的字段名
	FieldToTagName = "to"

	// 转换器标签

	// 特殊标签

	// 指定src结构体转换后的map的目标key
	// 对应的值为转换后的key名
	MapKeyTagName = "mk"
)

// 标签
type Tags map[string]string

// 标签池，优化内存分配
var tagsPool = sync.Pool{
	New: func() interface{} {
		return map[string]string{}
	},
}

// 清理标签池
func cleanTagsPool(kv map[string]string) {
	for k := range kv {
		delete(kv, k)
	}
	tagsPool.Put(kv)
}

// 获取字段上的标签
func getTags(field reflect.StructField) Tags {
	tags := tagsPool.Get().(map[string]string)

	for _, tag := range strings.Split(field.Tag.Get(TagName), TagSplitSymbol) {
		tagSlice := strings.SplitN(tag, TagValueSymbol, 2)

		switch len(tagSlice) {
		case 1:
			tags[tagSlice[0]] = ""
		default:
			tags[strings.TrimSpace(tagSlice[0])] = strings.TrimSpace(strings.Join(tagSlice[1:], ":"))
		}
	}

	return tags
}
