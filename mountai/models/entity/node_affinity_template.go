package entity

// 支持的节点亲和性 key
const (
	AffinityMatchExpressionKeySpec                   = "shanhai.int/spec"
	AffinityMatchExpressionKeyCPU                    = "shanhai.int/instance-cpu"
	AffinityMatchExpressionKeyMemory                 = "shanhai.int/instance-mem"
	AffinityMatchExpressionKeyExclusiveForDeployment = "shanhai.int/exclusive-deployment"
	AffinityMatchExpressionKeyExclusiveForJob        = "shanhai.int/exclusive-job"
)

// 亲和性匹配条件枚举值
const (
	AffinityMatchExpressionOperatorIn           = "In"
	AffinityMatchExpressionOperatorNotIn        = "NotIn"
	AffinityMatchExpressionOperatorExists       = "Exists"
	AffinityMatchExpressionOperatorDoesNotExist = "DoesNotExist"
	AffinityMatchExpressionOperatorGt           = "Gt"
	AffinityMatchExpressionOperatorLt           = "Lt"
)

// NodeAffinityTemplate : 节点亲和性配置
// 目前暂时只支持 requiredDuringSchedulingIgnoredDuringExecution
type NodeAffinityTemplate struct {
	RequiredDuringSchedulingIgnoredDuringExecution []NodeSelectorTerms
}

// NodeSelectorTerms : 节点亲和性配置条件
type NodeSelectorTerms struct {
	MatchExpressions []AffinityMatchExpression
}

// AffinityMatchExpression : 亲和性匹配条件
type AffinityMatchExpression struct {
	Key      string
	Operator string
	Values   []string
}

// GenerateNodeAffinityTemplate : 生成节点亲和性模板
func GenerateNodeAffinityTemplate(appType AppType, labelCfg NodeAffinityLabelConfig) NodeAffinityTemplate {
	expressions := make([]AffinityMatchExpression, 0, 16)
	// 添加服务分级(importance)标签表达式
	switch labelCfg.Importance {
	case ApplicationImportanceTypeLow:
		expressions = append(expressions, AffinityMatchExpression{
			Key:      AffinityMatchExpressionKeySpec,
			Operator: AffinityMatchExpressionOperatorIn,
			Values: []string{
				string(NodeLabelSpecTypeSmall),
			},
		})
	case ApplicationImportanceTypeSpecial:
		expressions = append(expressions, AffinityMatchExpression{
			Key:      AffinityMatchExpressionKeySpec,
			Operator: AffinityMatchExpressionOperatorIn,
			Values: []string{
				string(NodeLabelSpecTypeSpecial),
			},
		})
	case "", ApplicationImportanceTypeMedium: // 默认值为 medium
		expressions = append(expressions, AffinityMatchExpression{
			Key:      AffinityMatchExpressionKeySpec,
			Operator: AffinityMatchExpressionOperatorIn,
			Values: []string{
				string(NodeLabelSpecTypeSmall),
				string(NodeLabelSpecTypeMedium),
			},
		})

	case ApplicationImportanceTypeHigh:
		expressions = append(expressions, AffinityMatchExpression{
			Key:      AffinityMatchExpressionKeySpec,
			Operator: AffinityMatchExpressionOperatorIn,
			Values: []string{
				string(NodeLabelSpecTypeSmall),
				string(NodeLabelSpecTypeMedium),
				string(NodeLabelSpecTypeLarge),
			},
		})

	default:
	}

	// 添加CPU标签表达式
	if labelCfg.CPU == "" {
		expressions = append(expressions, AffinityMatchExpression{
			Key:      AffinityMatchExpressionKeyCPU,
			Operator: AffinityMatchExpressionOperatorDoesNotExist,
		})
	} else {
		expressions = append(expressions, AffinityMatchExpression{
			Key:      AffinityMatchExpressionKeyCPU,
			Operator: AffinityMatchExpressionOperatorIn,
			Values:   []string{string(labelCfg.CPU)},
		})
	}

	// 添加内存标签表达式
	if labelCfg.Mem == "" {
		expressions = append(expressions, AffinityMatchExpression{
			Key:      AffinityMatchExpressionKeyMemory,
			Operator: AffinityMatchExpressionOperatorDoesNotExist,
		})
	} else {
		expressions = append(expressions, AffinityMatchExpression{
			Key:      AffinityMatchExpressionKeyMemory,
			Operator: AffinityMatchExpressionOperatorIn,
			Values:   []string{string(labelCfg.Mem)},
		})
	}

	// 添加专用标签表达式
	expressions = AddExclusiveExpressions(appType, labelCfg.Exclusive, expressions)
	return NodeAffinityTemplate{
		RequiredDuringSchedulingIgnoredDuringExecution: []NodeSelectorTerms{
			{MatchExpressions: expressions},
		},
	}
}

// AddExclusiveExpressions : 添加专用标签表达式
func AddExclusiveExpressions(appType AppType, exclusive NodeLabelValueType,
	expressions []AffinityMatchExpression) []AffinityMatchExpression {
	switch appType {
	case AppTypeService, AppTypeWorker:
		if string(exclusive) == "" {
			expressions = append(expressions, AffinityMatchExpression{
				Key:      AffinityMatchExpressionKeyExclusiveForDeployment,
				Operator: AffinityMatchExpressionOperatorDoesNotExist,
			})
		} else {
			expressions = append(expressions, AffinityMatchExpression{
				Key:      AffinityMatchExpressionKeyExclusiveForDeployment,
				Operator: AffinityMatchExpressionOperatorIn,
				Values:   []string{string(exclusive)},
			})
		}

		expressions = append(expressions, AffinityMatchExpression{
			Key:      AffinityMatchExpressionKeyExclusiveForJob,
			Operator: AffinityMatchExpressionOperatorDoesNotExist,
		})

	case AppTypeCronJob:
		// CronJob 暂不支持专用标签
		expressions = append(expressions, AffinityMatchExpression{
			Key:      AffinityMatchExpressionKeyExclusiveForDeployment,
			Operator: AffinityMatchExpressionOperatorDoesNotExist,
		}, AffinityMatchExpression{
			Key:      AffinityMatchExpressionKeyExclusiveForJob,
			Operator: AffinityMatchExpressionOperatorDoesNotExist,
		})

	case AppTypeOneTimeJob:
		expressions = append(expressions, AffinityMatchExpression{
			Key:      AffinityMatchExpressionKeyExclusiveForDeployment,
			Operator: AffinityMatchExpressionOperatorDoesNotExist,
		})

		if string(exclusive) == "" {
			expressions = append(expressions, AffinityMatchExpression{
				Key:      AffinityMatchExpressionKeyExclusiveForDeployment,
				Operator: AffinityMatchExpressionOperatorDoesNotExist,
			})
		} else {
			expressions = append(expressions, AffinityMatchExpression{
				Key:      AffinityMatchExpressionKeyExclusiveForJob,
				Operator: AffinityMatchExpressionOperatorIn,
				Values:   []string{string(exclusive)},
			})
		}

	default:
		expressions = append(expressions, AffinityMatchExpression{
			Key:      AffinityMatchExpressionKeyExclusiveForDeployment,
			Operator: AffinityMatchExpressionOperatorDoesNotExist,
		}, AffinityMatchExpression{
			Key:      AffinityMatchExpressionKeyExclusiveForJob,
			Operator: AffinityMatchExpressionOperatorDoesNotExist,
		})
	}

	return expressions
}
