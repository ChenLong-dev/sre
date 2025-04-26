package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type isProjectImportanceLevelLabelTestcase struct {
	tag               string
	label             ProjectLabelValue
	isImportanceLabel bool
}

func Test_IsProjectImportanceLevelLabel(t *testing.T) {
	tcs := generateIsProjectImportanceLevelLabelTestcases()
	for _, tc := range tcs {
		curTc := tc
		t.Run(curTc.tag, func(t *testing.T) {
			assert.Equal(t, curTc.isImportanceLabel, curTc.label.IsProjectImportanceLevelLabel())
		})
	}
}

func generateIsProjectImportanceLevelLabelTestcases() []*isProjectImportanceLevelLabelTestcase {
	return []*isProjectImportanceLevelLabelTestcase{
		{
			tag:               "实际支持的 P0 级别",
			label:             ProjectLabelP0,
			isImportanceLabel: true,
		},
		{
			tag:               "实际支持的 P1 级别",
			label:             ProjectLabelP1,
			isImportanceLabel: true,
		},
		{
			tag:               "实际支持的 P2 级别",
			label:             ProjectLabelP2,
			isImportanceLabel: true,
		},
		{
			tag:               "多个零",
			label:             "P00",
			isImportanceLabel: true,
		},
		{
			tag:               "前导包含零",
			label:             "P01",
			isImportanceLabel: true,
		},
		{
			tag:               "超过 uint64",
			label:             "P18446744073709551616",
			isImportanceLabel: true,
		},
		{
			tag:               "实际支持的非级别标签",
			label:             ProjectLabelBff,
			isImportanceLabel: false,
		},
		{
			tag:               "实际支持的非级别标签",
			label:             ProjectLabelFrontend,
			isImportanceLabel: false,
		},
		{
			tag:               "实际支持的非级别标签",
			label:             ProjectLabelOnline,
			isImportanceLabel: false,
		},
		{
			tag:               "实际支持的非级别标签",
			label:             ProjectLabelOffline,
			isImportanceLabel: false,
		},
		{
			tag:               "实际支持的非级别标签",
			label:             ProjectLabelBackend,
			isImportanceLabel: false,
		},
		{
			tag:               "实际支持的非级别标签",
			label:             ProjectLabelInternal,
			isImportanceLabel: false,
		},
		{
			tag:               "实际支持的非级别标签",
			label:             ProjectLabelExternal,
			isImportanceLabel: false,
		},
		{
			tag:               "P开头非级别标签",
			label:             "Personal",
			isImportanceLabel: false,
		},
		{
			tag:               "纯数字",
			label:             "1",
			isImportanceLabel: false,
		},
	}
}
