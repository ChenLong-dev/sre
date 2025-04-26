package slice

func StrSliceIndexOf(strSlice []string, target string) int {
	for index, key := range strSlice {
		if key == target {
			return index
		}
	}

	return -1
}

// 判断字符串切片是否包含字符串
func StrSliceContains(strSlice []string, target string) bool {
	return StrSliceIndexOf(strSlice, target) != -1
}
