package utils

import (
	"reflect"
	"sort"
)

func SortedMapKeys(m interface{}) (keyList []string) {
	keys := reflect.ValueOf(m).MapKeys()

	for _, key := range keys {
		keyList = append(keyList, key.Interface().(string))
	}
	sort.Strings(keyList)
	return
}
