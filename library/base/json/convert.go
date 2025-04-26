package json

import (
	"reflect"
	"strings"
	"sync"
)

// 反射map对象池，优化内存分配
var tagsMapPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]interface{})
	},
}

func cleanTagsMap(kv map[string]interface{}) {
	if kv == nil {
		return
	}

	for k := range kv {
		delete(kv, k)
	}
	tagsMapPool.Put(kv)
}

// 获取json标签
func getJsonTags(field reflect.StructField) map[string]interface{} {
	tagString := field.Tag.Get("json")
	if tagString == "-" {
		return nil
	}

	tags := tagsMapPool.Get().(map[string]interface{})

	for idx, tag := range strings.Split(tagString, ",") {
		if idx == 0 && tag != "" {
			tags["column"] = tag
		} else {
			tags[tag] = true
		}
	}

	return tags
}

// 结构体转json可读取的map结构
// 暂时不建议使用
func StructToJsonMap(obj interface{}) map[string]interface{} {
	v := reflect.Indirect(reflect.ValueOf(obj))
	t := v.Type()

	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		var (
			vField = v.Field(i)
			tField = v.Type().Field(i)
			tagMap = getJsonTags(tField)
		)

		if tField.Type.Kind() == reflect.Struct && tField.Anonymous {
			childMap := StructToJsonMap(vField.Interface())
			for childKey, childValue := range childMap {
				data[childKey] = childValue
			}
			continue
		} else if tagMap == nil {
			continue
		} else if _, ok := tagMap["omitempty"]; vField.IsZero() && ok {
			continue
		} else if columnName, ok := tagMap["column"]; ok {
			data[columnName.(string)] = vField.Interface()
		} else {
			data[tField.Name] = vField.Interface()
		}

		cleanTagsMap(tagMap)
	}
	return data
}
