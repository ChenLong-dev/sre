package test

import (
	"testing"
)

// 同步测试单元
type SyncUnitCase struct {
	// 单测名称
	Name string
	// 单测函数
	Func func(t *testing.T)
}

// 同步运行用例
func RunSyncCases(t *testing.T, cases ...SyncUnitCase) {
	for _, c := range cases {
		t.Run(c.Name, c.Func)
	}
}
