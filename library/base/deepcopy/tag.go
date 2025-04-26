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

// 标签优先级从上到下，最高为skip
const (
	// 忽略当前转换
	SkipTagName = "skip"

	// 从src结构体的指定方法进行转换
	// 要求方法首个参数为 map[string]interface{} 类型
	// 要求方法首个返回值与当前字段相同，若最后一个参数为error类型，则会在非空时，抛出
	// 对应的值为方法名
	FieldMethodTagName = "method"

	// 从指定src结构体字段名称转换
	// 对应的值为字段名
	FieldFromTagName = "from"

	// 指定当前结构体转换后的目标名称
	// 对应的值为字段名
	FieldToTagName = "to"

	// 从src结构体任意类型强制转换成interface
	ForceTagName = "force"

	// 从src结构体time类型格式化指定string类型
	// 对应的值为time的format格式
	TimeFormatTagName = "timeformat"

	// 若源结构体没有符合条件的字段，则调用目标结构体的指定方法
	// 若不存在指定方法，则直接赋值给目标字段，目前只允许int，float64，string，bool类型
	// 对应的值为方法名
	DefaultMethodName = "default"
)

// 标签
type Tags map[string]string

// 标签池，优化内存分配
var kvMapPool = sync.Pool{
	New: func() interface{} {
		return map[string]string{}
	},
}

func cleanKVMap(kv map[string]string) {
	for k := range kv {
		delete(kv, k)
	}
	kvMapPool.Put(kv)
}

// 字段标签池，优化内存分配
var fieldTagsPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]Tags)
	},
}

func cleanFieldTags(fieldTags map[string]Tags) {
	for k, v := range fieldTags {
		cleanKVMap(v)
		delete(fieldTags, k)
	}
	fieldTagsPool.Put(fieldTags)
}

// 获取字段上的标签
func getTags(field reflect.StructField) Tags {
	tagString := field.Tag.Get(TagName)

	tags := kvMapPool.Get().(map[string]string)

	for _, tag := range strings.Split(tagString, TagSplitSymbol) {
		tagSlice := strings.Split(tag, TagValueSymbol)

		switch len(tagSlice) {
		case 1:
			tags[tagSlice[0]] = ""
		default:
			tags[strings.TrimSpace(tagSlice[0])] = strings.TrimSpace(strings.Join(tagSlice[1:], ":"))
		}
	}

	return tags
}

// 反射获取标签map
// key为字段名，value为标签
func getTagsMap(value reflect.Value, fieldNames []string) map[string]Tags {
	m := fieldTagsPool.Get().(map[string]Tags)

	for _, fieldName := range fieldNames {
		fieldType, fieldFound := value.Type().FieldByName(fieldName)
		if !fieldFound {
			continue
		}

		m[fieldType.Name] = getTags(fieldType)
	}

	return m
}

// 获取 FieldToTagName 标签的映射map
// key为转换后的dst字段名，value为src字段名
func getFieldToTagMap(tagsMap map[string]Tags) map[string]string {
	var (
		toMap = kvMapPool.Get().(map[string]string)
	)

	for field, tags := range tagsMap {
		if toName, ok := tags[FieldToTagName]; ok && toName != "" {
			toMap[toName] = field
		}
	}

	return toMap
}
