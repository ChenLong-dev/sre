// k8s GroupVersion
// 文档: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#resources
// 主格式: {Group}/{Version} (其中 Group 允许为空, 此时整体版本不带'/', 例: v1)
// 主格式字段:
//
//	Group:
//	  域名格式([a-z\.], 最大长度253字符, '.'为分隔符不在首尾, 不连续), 官方推荐使用子域名格式
//	Version(字段名不全是官方取的, 有些是个人理解):
//	  DNS_LABEL格式([a-z0-9\-], 并且'-'不能在首尾)
//	  文档: https://github.com/kubernetes/community/blob/master/contributors/design-proposals/architecture/identifiers.md
//	    官方目前的 Version 格式(字段命名是个人理解): v{MainVersion}{TestPhase}{TestVersion}
//	    其中 MainVersion 必定包含, 实验版本部分不一定包含
//	      例1: apps/v1
//	        Group = apps
//	        MainVersion = 1
//	      例2: agent.k8s.elastic.co/v1alpha1
//	        Group = agent.k8s.elastic.co
//	        MainVersion = 1
//	        TestPhase = alpha
//	        TestVersion = 1
package entity

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"k8s.io/apimachinery/pkg/runtime/schema"

	_errcode "rulai/utils/errcode"
)

// K8sAPITestPhase k8s GroupVersion 实验阶段
type K8sAPITestPhase string

const (
	K8sAPITestPhaseEmpty K8sAPITestPhase = ""
	K8sAPITestPhaseAlpha K8sAPITestPhase = "alpha"
	K8sAPITestPhaseBeta  K8sAPITestPhase = "beta"
)

// GroupNameExtensions extensions 组, 该组在 k8s 的资源中一般不优先使用
const GroupNameExtensions = "extensions"

var allowedTestPhaseStrings = []string{string(K8sAPITestPhaseAlpha), string(K8sAPITestPhaseBeta)}

// k8sGroupVersionOfficialVersionRegex k8s GroupVersion.Version 官方版本的正则表达式
var k8sGroupVersionOfficialVersionRegex = regexp.MustCompile(
	fmt.Sprintf(`^v([0-9]+)((%s)([0-9]+)){0,1}$`, strings.Join(allowedTestPhaseStrings, "|")))

// StandardK8sGroupVersion 标准 k8s GroupVersion 结构
type StandardK8sGroupVersion struct {
	GroupVersion *schema.GroupVersion

	MainVersion int
	TestPhase   K8sAPITestPhase
	TestVersion int
}

func (ver *StandardK8sGroupVersion) isValid() bool {
	if ver.GroupVersion == nil || ver.MainVersion <= 0 {
		return false
	}

	switch ver.TestPhase {
	case K8sAPITestPhaseEmpty:
		return ver.TestVersion == -1 && ver.GroupVersion.Version == fmt.Sprintf("v%d", ver.MainVersion)

	case K8sAPITestPhaseAlpha, K8sAPITestPhaseBeta:
		return ver.TestPhase != K8sAPITestPhaseEmpty &&
			ver.GroupVersion.Version == fmt.Sprintf("v%d%s%d", ver.MainVersion, ver.TestPhase, ver.TestVersion)

	default:
	}

	return false
}

// IsPreferredThan 比较当前版本在 AMS 系统是否比目标版本更有倾向性
func (ver *StandardK8sGroupVersion) IsPreferredThan(target *schema.GroupVersion) (bool, error) {
	if !ver.isValid() {
		return false, errors.Wrapf(errcode.InternalError, "invalid StandardK8sGroupVersion(%v)", ver)
	}

	if target == nil {
		return true, nil
	}

	targetVer, err := parseK8sGroupVersion(target)
	if err != nil {
		return false, err
	}

	// extensions 组的 API 版本倾向性低于其他正式组(包括空组)
	if target.Group == GroupNameExtensions && ver.GroupVersion.Group != GroupNameExtensions {
		return true, nil
	}

	if target.Group != GroupNameExtensions && ver.GroupVersion.Group == GroupNameExtensions {
		return false, nil
	}

	if ver.GroupVersion.Group != target.Group {
		// 暂时没有除了 extensions 组和 空组 之外的有多组的情况, 如果未来遇到需要再看如何处理
		return false, errors.Wrapf(errcode.InternalError,
			"cannot compare with different groups(%s, %s)", ver.GroupVersion.Group, target.Group)
	}

	if ver.MainVersion == targetVer.MainVersion {
		switch ver.TestPhase {
		case K8sAPITestPhaseEmpty:
			return targetVer.TestPhase != K8sAPITestPhaseEmpty, nil

		case K8sAPITestPhaseAlpha:
			return targetVer.TestPhase == K8sAPITestPhaseAlpha && ver.TestVersion > targetVer.TestVersion, nil

		case K8sAPITestPhaseBeta:
			return targetVer.TestPhase == K8sAPITestPhaseAlpha ||
				(targetVer.TestPhase == K8sAPITestPhaseBeta && ver.TestVersion > targetVer.TestVersion), nil

		default:
			// 检测时已处理未知 phase
		}
	}

	return ver.MainVersion > targetVer.MainVersion, nil
}

// ParseK8sGroupVersionString 解析 k8s GroupVersion 字符串
func ParseK8sGroupVersionString(gv string) (*StandardK8sGroupVersion, error) {
	groupVersion, err := schema.ParseGroupVersion(gv)
	if err != nil {
		return nil, errors.Wrapf(_errcode.K8sInternalError, "parse GroupVersion error: %s", err)
	}

	return parseK8sGroupVersion(&groupVersion)
}

// parseK8sGroupVersion 解析 k8s schema.GroupVersion 的版本
func parseK8sGroupVersion(groupVersion *schema.GroupVersion) (*StandardK8sGroupVersion, error) {
	parts := k8sGroupVersionOfficialVersionRegex.FindStringSubmatch(groupVersion.Version)
	// 正则匹配结果是: [{Version}, {MainVersion}, {TestPhaseAndVersion}, {TestPhase}, {TestVersion}]
	if len(parts) != 5 {
		// 未匹配
		return nil, errors.Wrapf(errcode.InternalError, "unknown version format of GroupVersion(%s)", groupVersion.String())
	}

	ver := &StandardK8sGroupVersion{GroupVersion: groupVersion}
	// 正则已经保证 MainVersion 可转化为非负整数, 只需要判断零值
	if ver.MainVersion, _ = strconv.Atoi(parts[1]); ver.MainVersion < 1 {
		return nil, errors.Wrapf(errcode.InternalError, "unknown main_version in GroupVersion(%s)", groupVersion.String())
	}

	// 正则已经保证 TestVersion 可转化为非负整数, 只需要判断零值
	if parts[4] == "" {
		ver.TestVersion = -1 // 避免未来出现 TestVersion=0 的情形
	} else if ver.TestVersion, _ = strconv.Atoi(parts[4]); ver.TestVersion < 1 {
		return nil, errors.Wrapf(errcode.InternalError, "unknown test_version in GroupVersion(%s)", groupVersion.String())
	}

	ver.TestPhase = K8sAPITestPhase(parts[3]) // 正则已经保证了这里是合法的 phase, 如果没有 TestPhase 则为空值

	return ver, nil
}
