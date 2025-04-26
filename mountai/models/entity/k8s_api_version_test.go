package entity

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type parseK8sAPIVersionTestcase struct {
	tag    string
	gv     string
	ver    *StandardK8sGroupVersion
	errStr string
}

type preferredK8sAPIVersionTestcase struct {
	tag         string
	current     *StandardK8sGroupVersion
	target      *schema.GroupVersion
	isPreferred bool
	errStr      string
}

func Test_ParseK8sAPIVersion(t *testing.T) {
	tcs := generateParseK8sAPIVersionTestcases()
	for _, tc := range tcs {
		currentTc := tc
		t.Run(currentTc.tag, func(t *testing.T) {
			ver, err := ParseK8sGroupVersionString(currentTc.gv)
			if currentTc.errStr != "" {
				assert.EqualError(t, err, currentTc.errStr)
				assert.Nil(t, ver)
				return
			}

			assert.Nil(t, err)
			assert.EqualValues(t, currentTc.ver, ver)
			assert.Equal(t, currentTc.gv, ver.GroupVersion.String())
		})
	}
}

func Test_K8sAPIVersion_IsPreferredThan(t *testing.T) {
	tcs := generatePreferredK8sAPIVersionTestcases(t)
	for _, tc := range tcs {
		currentTc := tc
		t.Run(currentTc.tag, func(t *testing.T) {
			isPreferred, err := currentTc.current.IsPreferredThan(currentTc.target)
			assert.Equal(t, currentTc.isPreferred, isPreferred)

			if currentTc.errStr == "" {
				assert.Nil(t, err)
			} else {
				assert.EqualError(t, err, currentTc.errStr)
			}
		})
	}
}

func generateParseK8sAPIVersionTestcases() []*parseK8sAPIVersionTestcase {
	return []*parseK8sAPIVersionTestcase{
		{
			tag:    "[异常] 空值",
			gv:     "",
			errStr: "unknown version format of GroupVersion(): 1050500:系统错误,请稍后重试",
		},
		{
			tag:    "[异常] 2个斜杠",
			gv:     "head/middle/tail",
			errStr: "parse GroupVersion error: unexpected GroupVersion string: head/middle/tail: 9010007:K8s内部错误",
		},
		{
			tag:    "[异常] 多个斜杠",
			gv:     "head/middle1/middle2/middle3/tail",
			errStr: "parse GroupVersion error: unexpected GroupVersion string: head/middle1/middle2/middle3/tail: 9010007:K8s内部错误",
		},
		{
			tag:    "[异常] 无 Group, Version 完全不是官方格式",
			gv:     "unknown_version",
			errStr: "unknown version format of GroupVersion(unknown_version): 1050500:系统错误,请稍后重试",
		},
		{
			tag:    "[异常] 无 Group, MainVersion 无法匹配",
			gv:     "v_1alpha1",
			errStr: "unknown version format of GroupVersion(v_1alpha1): 1050500:系统错误,请稍后重试",
		},
		{
			tag:    "[异常] 无 Group, TestPhase 无法匹配",
			gv:     "v1delta1",
			errStr: "unknown version format of GroupVersion(v1delta1): 1050500:系统错误,请稍后重试",
		},
		{
			tag:    "[异常] 无 Group, TestVersion 无法匹配",
			gv:     "v1alpha_1",
			errStr: "unknown version format of GroupVersion(v1alpha_1): 1050500:系统错误,请稍后重试",
		},
		{
			tag:    "[异常] 无 Group, 多出前缀",
			gv:     "unknown_prefix-v1alpha1",
			errStr: "unknown version format of GroupVersion(unknown_prefix-v1alpha1): 1050500:系统错误,请稍后重试",
		},
		{
			tag:    "[异常] 无 Group, 多出后缀",
			gv:     "v1alpha1-unknown_suffix",
			errStr: "unknown version format of GroupVersion(v1alpha1-unknown_suffix): 1050500:系统错误,请稍后重试",
		},
		{
			tag:    "[异常] Group=unittest.qtfm.cn, Version 完全不是官方格式",
			gv:     "unittest.qtfm.cn/unknown_version",
			errStr: "unknown version format of GroupVersion(unittest.qtfm.cn/unknown_version): 1050500:系统错误,请稍后重试",
		},
		{
			tag:    "[异常] Group=unittest.qtfm.cn, MainVersion 无法匹配",
			gv:     "unittest.qtfm.cn/v_1alpha1",
			errStr: "unknown version format of GroupVersion(unittest.qtfm.cn/v_1alpha1): 1050500:系统错误,请稍后重试",
		},
		{
			tag:    "[异常] Group=unittest.qtfm.cn, TestPhase 无法匹配",
			gv:     "unittest.qtfm.cn/v1delta1",
			errStr: "unknown version format of GroupVersion(unittest.qtfm.cn/v1delta1): 1050500:系统错误,请稍后重试",
		},
		{
			tag:    "[异常] Group=unittest.qtfm.cn, TestVersion 无法匹配",
			gv:     "unittest.qtfm.cn/v1alpha_1",
			errStr: "unknown version format of GroupVersion(unittest.qtfm.cn/v1alpha_1): 1050500:系统错误,请稍后重试",
		},
		{
			tag:    "[异常] Group=unittest.qtfm.cn, 多出前缀",
			gv:     "unittest.qtfm.cn/unknown_prefix-v1alpha1",
			errStr: "unknown version format of GroupVersion(unittest.qtfm.cn/unknown_prefix-v1alpha1): 1050500:系统错误,请稍后重试",
		},
		{
			tag:    "[异常] Group=unittest.qtfm.cn, 多出后缀",
			gv:     "unittest.qtfm.cn/v1alpha1-unknown_suffix",
			errStr: "unknown version format of GroupVersion(unittest.qtfm.cn/v1alpha1-unknown_suffix): 1050500:系统错误,请稍后重试",
		},
		{
			tag:    "[异常] Group=unittest.qtfm.cn, MainVersion=0",
			gv:     "unittest.qtfm.cn/v0alpha1",
			errStr: "unknown main_version in GroupVersion(unittest.qtfm.cn/v0alpha1): 1050500:系统错误,请稍后重试",
		},
		{
			tag:    "[异常] Group=unittest.qtfm.cn, TestVersion=0",
			gv:     "unittest.qtfm.cn/v1alpha0",
			errStr: "unknown test_version in GroupVersion(unittest.qtfm.cn/v1alpha0): 1050500:系统错误,请稍后重试",
		},
		{
			tag: "[正常] 无 Group, MainVersion=1",
			gv:  "v1",
			ver: &StandardK8sGroupVersion{
				GroupVersion: &schema.GroupVersion{
					Version: "v1",
				},
				MainVersion: 1,
				TestVersion: -1,
			},
		},
		{
			tag: "[正常] 无 Group, MainVersion=1, TestPhase=alpha, TestVersion=1",
			gv:  "v1alpha1",
			ver: &StandardK8sGroupVersion{
				GroupVersion: &schema.GroupVersion{
					Version: "v1alpha1",
				},
				MainVersion: 1,
				TestPhase:   K8sAPITestPhaseAlpha,
				TestVersion: 1,
			},
		},
		{
			tag: "[正常] 无 Group, MainVersion=2, TestPhase=beta, TestVersion=2",
			gv:  "v2beta2",
			ver: &StandardK8sGroupVersion{
				GroupVersion: &schema.GroupVersion{
					Version: "v2beta2",
				},
				MainVersion: 2,
				TestPhase:   K8sAPITestPhaseBeta,
				TestVersion: 2,
			},
		},
		{
			tag: "[正常] Group=unittest.qtfm.cn, MainVersion=1",
			gv:  "unittest.qtfm.cn/v1",
			ver: &StandardK8sGroupVersion{
				GroupVersion: &schema.GroupVersion{
					Group:   "unittest.qtfm.cn",
					Version: "v1",
				},
				MainVersion: 1,
				TestVersion: -1,
			},
		},
		{
			tag: "[正常] Group=unittest.qtfm.cn, MainVersion=1, TestPhase=alpha, TestVersion=1",
			gv:  "unittest.qtfm.cn/v1alpha1",
			ver: &StandardK8sGroupVersion{
				GroupVersion: &schema.GroupVersion{
					Group:   "unittest.qtfm.cn",
					Version: "v1alpha1",
				},
				MainVersion: 1,
				TestPhase:   K8sAPITestPhaseAlpha,
				TestVersion: 1,
			},
		},
		{
			tag: "[正常] Group=unittest.qtfm.cn, MainVersion=2, TestPhase=beta, TestVersion=2",
			gv:  "unittest.qtfm.cn/v2beta2",
			ver: &StandardK8sGroupVersion{
				GroupVersion: &schema.GroupVersion{
					Group:   "unittest.qtfm.cn",
					Version: "v2beta2",
				},
				MainVersion: 2,
				TestPhase:   K8sAPITestPhaseBeta,
				TestVersion: 2,
			},
		},
	}
}

func generatePreferredK8sAPIVersionTestcases(t *testing.T) []*preferredK8sAPIVersionTestcase {
	commonValidGroupVersion := &schema.GroupVersion{
		Group:   GroupNameExtensions,
		Version: "v1alpha1",
	}
	commonValidStandardK8sGroupVersion, err := parseK8sGroupVersion(commonValidGroupVersion)
	require.Nil(t, err, "prepare commonValidGroupVersion: %v", commonValidGroupVersion)

	invalidStandardK8sGroupVersions := map[string]*StandardK8sGroupVersion{
		"[异常] current GroupVersion == nil": {
			GroupVersion: nil,
			MainVersion:  1,
			TestVersion:  -1,
		},
		"[异常] current MainVersion == 0": {
			GroupVersion: &schema.GroupVersion{Version: "v0"},
			MainVersion:  0,
			TestVersion:  -1,
		},
		"[异常] current TestPhase 非法": {
			GroupVersion: &schema.GroupVersion{
				Version: "v1delta1",
			},
			MainVersion: 1,
			TestPhase:   K8sAPITestPhase("delta"),
			TestVersion: 1,
		},
		"[异常] current TestPhase 为空, TestVersion 有正常值": {
			GroupVersion: &schema.GroupVersion{Version: "v13"},
			MainVersion:  1,
			TestVersion:  3,
		},
		"[异常] current TestPhase 为空, TestVersion 非法": {
			GroupVersion: &schema.GroupVersion{Version: "v1-2"},
			MainVersion:  1,
			TestVersion:  -2,
		},
		"[异常] current TestPhase 非空, TestVersion == 0": {
			GroupVersion: &schema.GroupVersion{Version: "v1alpha"},
			MainVersion:  1,
			TestPhase:    K8sAPITestPhaseAlpha,
		},
		"[异常] current TestPhase 非空, TestVersion == -1": {
			GroupVersion: &schema.GroupVersion{Version: "v1alpha"},
			MainVersion:  1,
			TestPhase:    K8sAPITestPhaseAlpha,
			TestVersion:  -1,
		},
		"[异常] current TestPhase 非空, TestVersion 非法": {
			GroupVersion: &schema.GroupVersion{Version: "v1alpha"},
			MainVersion:  1,
			TestPhase:    K8sAPITestPhaseAlpha,
			TestVersion:  -2,
		},
	}

	invalidGroupVersions := map[string]*schema.GroupVersion{
		"[异常] target TestPhase 非法": {Version: "v1delta1"},
	}

	var tcs []*preferredK8sAPIVersionTestcase
	for tag, ver := range invalidStandardK8sGroupVersions {
		tcs = append(tcs, &preferredK8sAPIVersionTestcase{
			tag:     tag,
			current: ver,
			target:  commonValidGroupVersion,
			errStr:  fmt.Sprintf("invalid StandardK8sGroupVersion(%v): 1050500:系统错误,请稍后重试", ver),
		})
	}

	for tag, ver := range invalidGroupVersions {
		tcs = append(tcs, &preferredK8sAPIVersionTestcase{
			tag:     tag,
			current: commonValidStandardK8sGroupVersion,
			target:  ver,
			errStr:  fmt.Sprintf("unknown version format of GroupVersion(%s): 1050500:系统错误,请稍后重试", ver),
		})
	}

	items := generatePreferredK8sAPIVersionTestcaseItem()
	// i 用于遍历 target, j 用于遍历 current, 注意 current 不能为空
	for i := range items {
		for j := range items {
			if items[j] == nil {
				continue
			}

			tc := &preferredK8sAPIVersionTestcase{target: items[i]}

			tc.current, err = parseK8sGroupVersion(items[j])
			require.Nil(t, err, tc.tag)

			if tc.target != nil &&
				tc.current.GroupVersion.Group != tc.target.Group &&
				tc.current.GroupVersion.Group != GroupNameExtensions &&
				tc.target.Group != GroupNameExtensions {
				tc.errStr = fmt.Sprintf(
					"cannot compare with different groups(%s, %s): 1050500:系统错误,请稍后重试",
					tc.current.GroupVersion.Group, tc.target.Group)
				tc.tag = fmt.Sprintf("[正常] current(%s), target(%s), with error: %s", items[j], items[i], tc.errStr)
			} else {
				tc.isPreferred = j > i
				tc.tag = fmt.Sprintf("[正常] current(%s), target(%s), is_preferred: %t", items[j], items[i], tc.isPreferred)
			}

			tcs = append(tcs, tc)
		}
	}

	return tcs
}

// generatePreferredK8sAPIVersionTestcaseItem 按照优先级自动构建测试用例(靠前的优先级低)
func generatePreferredK8sAPIVersionTestcaseItem() []*schema.GroupVersion {
	return []*schema.GroupVersion{
		nil,
		{
			Group:   GroupNameExtensions,
			Version: "v1alpha1",
		},
		{
			Group:   GroupNameExtensions,
			Version: "v1alpha2",
		},
		{
			Group:   GroupNameExtensions,
			Version: "v1beta1",
		},
		{
			Group:   GroupNameExtensions,
			Version: "v1beta2",
		},
		{
			Group:   GroupNameExtensions,
			Version: "v1",
		},
		{
			Group:   GroupNameExtensions,
			Version: "v2alpha1",
		},
		{
			Group:   GroupNameExtensions,
			Version: "v2alpha2",
		},
		{
			Group:   GroupNameExtensions,
			Version: "v2beta1",
		},
		{
			Group:   GroupNameExtensions,
			Version: "v2beta2",
		},
		{
			Group:   GroupNameExtensions,
			Version: "v2",
		},
		{
			Group:   "unittest.qtfm.cn",
			Version: "v1alpha1",
		},
		{
			Group:   "unittest.qtfm.cn",
			Version: "v1alpha2",
		},
		{
			Group:   "unittest.qtfm.cn",
			Version: "v1beta1",
		},
		{
			Group:   "unittest.qtfm.cn",
			Version: "v1beta2",
		},
		{
			Group:   "unittest.qtfm.cn",
			Version: "v1",
		},
		{
			Group:   "unittest.qtfm.cn",
			Version: "v2alpha1",
		},
		{
			Group:   "unittest.qtfm.cn",
			Version: "v2alpha2",
		},
		{
			Group:   "unittest.qtfm.cn",
			Version: "v2beta1",
		},
		{
			Group:   "unittest.qtfm.cn",
			Version: "v2beta2",
		},
		{
			Group:   "unittest.qtfm.cn",
			Version: "v2",
		},
		{
			Version: "v1alpha1",
		},
		{
			Version: "v1alpha2",
		},
		{
			Version: "v1beta1",
		},
		{
			Version: "v1beta2",
		},
		{
			Version: "v1",
		},
		{
			Version: "v2alpha1",
		},
		{
			Version: "v2alpha2",
		},
		{
			Version: "v2beta1",
		},
		{
			Version: "v2beta2",
		},
		{
			Version: "v2",
		},
	}
}
