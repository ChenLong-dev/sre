package decimal

import (
	"fmt"
	"strconv"
)

// 用于保留对应位数小数
// 	digit为保留位数，number为小数
func ToFixed(digit int, number float64) (float64, error) {
	return strconv.ParseFloat(
		fmt.Sprintf("%."+strconv.Itoa(digit)+"f", number),
		64,
	)
}
