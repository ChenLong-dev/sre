package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gopkg.in/yaml.v3"

	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
)

type labelTestcase struct {
	tag             string
	label           entity.NodeAffinityLabelConfig
	baseTemplateCfg entity.NodeAffinityTemplate
}

type nodeSelectorTestcase struct {
	tag          string
	nodeSelector map[string]string
	expressions  []entity.AffinityMatchExpression
}

func fillTemplateCfg(appType entity.AppType, exclusiveValue entity.NodeLabelValueType, tpl entity.NodeAffinityTemplate,
	oldExpressions []entity.AffinityMatchExpression) entity.NodeAffinityTemplate {
	expressions := make([]entity.AffinityMatchExpression, 0, 16)

	for _, exp := range tpl.RequiredDuringSchedulingIgnoredDuringExecution[0].MatchExpressions {
		expressions = append(expressions, entity.AffinityMatchExpression{
			Key:      exp.Key,
			Operator: exp.Operator,
			Values:   exp.Values,
		})
	}

	expressions = entity.AddExclusiveExpressions(appType, exclusiveValue, expressions)
	newTpl := entity.NodeAffinityTemplate{
		RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
			{MatchExpressions: expressions},
		},
	}

	if len(oldExpressions) > 0 {
		newTpl.RequiredDuringSchedulingIgnoredDuringExecution = append(
			newTpl.RequiredDuringSchedulingIgnoredDuringExecution,
			entity.NodeSelectorTerms{MatchExpressions: oldExpressions},
		)
	}
	return newTpl
}

func testLabelAndNodeSelector(t *testing.T, labelTestcases []labelTestcase, nodeSelectorTestcases []nodeSelectorTestcase,
	appType entity.AppType, startTask *resp.TaskDetailResp) {
	for _, clusterSet := range s.k8sClusters {
		for clusterName := range clusterSet {
			for _, ltcTmp := range labelTestcases {
				ltc := ltcTmp
				for _, ntcTmp := range nodeSelectorTestcases {
					ntc := ntcTmp

					// CronJob 跳过 exclusive 测试
					if appType == entity.AppTypeCronJob && ltc.label.Exclusive != "" {
						continue
					}

					ensuredClusterName := clusterName
					tag := fmt.Sprintf("%s with %s and %s in cluster(%s)", startTask.Version, ltc.tag, ntc.tag, clusterName)
					t.Run(
						tag,
						func(t *testing.T) {
							task := new(resp.TaskDetailResp)
							err := deepcopy.Copy(startTask).To(task)
							assert.NoError(t, err)

							tplCfg := fillTemplateCfg(
								appType,
								ltc.label.Exclusive,
								ltc.baseTemplateCfg,
								ntc.expressions)
							task.Param.NodeAffinityLabelConfig = ltc.label
							task.Param.NodeSelector = ntc.nodeSelector
							task.ClusterName = ensuredClusterName

							var (
								tpl          entity.K8sObjectTemplate
								nodeAffinity entity.NodeAffinityTemplate
								tolerations  map[string]string
								tplPath      string
							)

							if appType == entity.AppTypeCronJob {
								cronJobTpl, e := s.initCronJobTemplate(
									context.Background(),
									testProject,
									testCronJobApp,
									task,
									testTeam,
								)
								require.NoError(t, e)
								nodeAffinity = cronJobTpl.NodeAffinity
								tolerations = cronJobTpl.Tolerations
								tpl = cronJobTpl
								tplPath = testTemplateFileDir
							} else if appType == entity.AppTypeOneTimeJob {
								jobTpl, e := s.initJobTemplate(
									context.Background(),
									testProject,
									testJobApp,
									task,
									testTeam,
								)
								require.NoError(t, e)

								nodeAffinity = jobTpl.NodeAffinity
								tolerations = jobTpl.Tolerations
								tpl = jobTpl
								tplPath = testTemplateFileDir
							} else {
								deployTpl, e := s.initDeploymentTemplate(
									context.Background(),
									testProject,
									testRestfulServiceApp,
									task,
									testTeam,
								)
								require.NoError(t, e)
								nodeAffinity = deployTpl.NodeAffinity
								tolerations = deployTpl.Tolerations
								tpl = deployTpl
								tplPath = testTemplateFileDir
							}

							assert.Len(
								t,
								nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
								len(tplCfg.RequiredDuringSchedulingIgnoredDuringExecution),
								tag)
							for i := range tplCfg.RequiredDuringSchedulingIgnoredDuringExecution {
								assert.ElementsMatch(
									t,
									tplCfg.RequiredDuringSchedulingIgnoredDuringExecution[i].MatchExpressions,
									nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution[i].MatchExpressions,
									tag)
							}
							assert.EqualValues(t, ntc.nodeSelector, tolerations, tag)

							data, err := s.RenderK8sTemplate(
								context.Background(), tplPath, ensuredClusterName, entity.AppEnvStg, tpl)
							assert.NoError(t, err)
							fmt.Printf("%s\n", data)

							res := make(map[string]interface{})
							err = yaml.Unmarshal([]byte(data), &res)
							assert.NoError(t, err)
							fmt.Printf("%v\n", res)
						},
					)
				}
			}
		}
	}
}

func generateNodeSelectorTestcases() []nodeSelectorTestcase {
	testNilNodeSelectorTestcase := nodeSelectorTestcase{
		tag: "nil NodeSelectors",
	}
	testEmptyNodeSelectorTestcase := nodeSelectorTestcase{
		tag:          "empty NodeSelectors",
		nodeSelector: map[string]string{},
	}
	testSingleNodeSelectorsTestcase := nodeSelectorTestcase{
		tag: "single NodeSelectors",
		nodeSelector: map[string]string{
			"test_key_1": "test_value_1",
		},
		expressions: []entity.AffinityMatchExpression{
			{
				Key:      "test_key_1",
				Operator: entity.AffinityMatchExpressionOperatorIn,
				Values:   []string{"test_value_1"},
			},
		},
	}
	testMultipleNodeSelectorsTestcase := nodeSelectorTestcase{
		tag: "multiple NodeSelectors",
		nodeSelector: map[string]string{
			"test_key_2": "test_value_2",
			"test_key_3": "test_value_3",
			"test_key_4": "test_value_4",
		},
		expressions: []entity.AffinityMatchExpression{
			{
				Key:      "test_key_2",
				Operator: entity.AffinityMatchExpressionOperatorIn,
				Values:   []string{"test_value_2"},
			},
			{
				Key:      "test_key_3",
				Operator: entity.AffinityMatchExpressionOperatorIn,
				Values:   []string{"test_value_3"},
			},
			{
				Key:      "test_key_4",
				Operator: entity.AffinityMatchExpressionOperatorIn,
				Values:   []string{"test_value_4"},
			},
		},
	}

	return []nodeSelectorTestcase{
		testNilNodeSelectorTestcase,
		testEmptyNodeSelectorTestcase,
		testSingleNodeSelectorsTestcase,
		testMultipleNodeSelectorsTestcase,
	}
}

func generateLabelTestcases() []labelTestcase {
	testCPULabelValue := entity.NodeLabelValueType("test_cpu_type-8")
	testMemoryLabelValue := entity.NodeLabelValueType("test_mem_type-16G")
	testExclusiveLabelValue := entity.NodeLabelValueType("test_exclusive_tag")

	return []labelTestcase{
		{
			tag: "importance=low",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeLow,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeLow),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=medium",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeMedium,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeLow),
									string(entity.NodeLabelImportanceTypeMedium),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=high",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeHigh,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeLow),
									string(entity.NodeLabelImportanceTypeMedium),
									string(entity.NodeLabelImportanceTypeHigh),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=special",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeSpecial,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeSpecial),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=low,cpu",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeLow,
				CPU:        testCPULabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeLow),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testCPULabelValue)},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=medium,cpu",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeMedium,
				CPU:        testCPULabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeLow),
									string(entity.NodeLabelImportanceTypeMedium),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testCPULabelValue)},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=high,cpu",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeHigh,
				CPU:        testCPULabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeLow),
									string(entity.NodeLabelImportanceTypeMedium),
									string(entity.NodeLabelImportanceTypeHigh),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testCPULabelValue)},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=special,cpu",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeSpecial,
				CPU:        testCPULabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeSpecial),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testCPULabelValue)},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=low,mem",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeLow,
				Mem:        testMemoryLabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeLow),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testMemoryLabelValue)},
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=medium,mem",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeMedium,
				Mem:        testMemoryLabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeLow),
									string(entity.NodeLabelImportanceTypeMedium),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testMemoryLabelValue)},
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=high,mem",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeHigh,
				Mem:        testMemoryLabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeLow),
									string(entity.NodeLabelImportanceTypeMedium),
									string(entity.NodeLabelImportanceTypeHigh),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testMemoryLabelValue)},
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=special,mem",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeSpecial,
				Mem:        testMemoryLabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeSpecial),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testMemoryLabelValue)},
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=low,exclusive",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeLow,
				Exclusive:  testExclusiveLabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeLow),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=medium,exclusive",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeMedium,
				Exclusive:  testExclusiveLabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeLow),
									string(entity.NodeLabelImportanceTypeMedium),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=high,exclusive",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeHigh,
				Exclusive:  testExclusiveLabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeLow),
									string(entity.NodeLabelImportanceTypeMedium),
									string(entity.NodeLabelImportanceTypeHigh),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=special,exclusive",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeSpecial,
				Exclusive:  testExclusiveLabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeSpecial),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=low,cpu,mem",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeLow,
				CPU:        testCPULabelValue,
				Mem:        testMemoryLabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeLow),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testCPULabelValue)},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testMemoryLabelValue)},
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=medium,cpu,mem",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeMedium,
				CPU:        testCPULabelValue,
				Mem:        testMemoryLabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeLow),
									string(entity.NodeLabelImportanceTypeMedium),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testCPULabelValue)},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testMemoryLabelValue)},
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=high,cpu,mem",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeHigh,
				CPU:        testCPULabelValue,
				Mem:        testMemoryLabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeLow),
									string(entity.NodeLabelImportanceTypeMedium),
									string(entity.NodeLabelImportanceTypeHigh),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testCPULabelValue)},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testMemoryLabelValue)},
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=special,cpu,mem",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeSpecial,
				CPU:        testCPULabelValue,
				Mem:        testMemoryLabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeSpecial),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testCPULabelValue)},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testMemoryLabelValue)},
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=low,cpu,exclusive",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeLow,
				CPU:        testCPULabelValue,
				Exclusive:  testExclusiveLabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeLow),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testCPULabelValue)},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=medium,cpu,exclusive",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeMedium,
				CPU:        testCPULabelValue,
				Exclusive:  testExclusiveLabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeLow),
									string(entity.NodeLabelImportanceTypeMedium),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testCPULabelValue)},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=high,cpu,exclusive",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeHigh,
				CPU:        testCPULabelValue,
				Exclusive:  testExclusiveLabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeLow),
									string(entity.NodeLabelImportanceTypeMedium),
									string(entity.NodeLabelImportanceTypeHigh),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testCPULabelValue)},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=special,cpu,exclusive",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeSpecial,
				CPU:        testCPULabelValue,
				Exclusive:  testExclusiveLabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeSpecial),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testCPULabelValue)},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=low,mem,exclusive",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeLow,
				Mem:        testMemoryLabelValue,
				Exclusive:  testExclusiveLabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeLow),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testMemoryLabelValue)},
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=medium,mem,exclusive",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeMedium,
				Mem:        testMemoryLabelValue,
				Exclusive:  testExclusiveLabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeLow),
									string(entity.NodeLabelImportanceTypeMedium),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testMemoryLabelValue)},
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=high,mem,exclusive",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeHigh,
				Mem:        testMemoryLabelValue,
				Exclusive:  testExclusiveLabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeLow),
									string(entity.NodeLabelImportanceTypeMedium),
									string(entity.NodeLabelImportanceTypeHigh),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testMemoryLabelValue)},
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=special,mem,exclusive",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeSpecial,
				Mem:        testMemoryLabelValue,
				Exclusive:  testExclusiveLabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeSpecial),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorDoesNotExist,
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testMemoryLabelValue)},
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=low,cpu,mem,exclusive",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeLow,
				CPU:        testCPULabelValue,
				Mem:        testMemoryLabelValue,
				Exclusive:  testExclusiveLabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeLow),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testCPULabelValue)},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testMemoryLabelValue)},
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=medium,cpu,mem,exclusive",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeMedium,
				CPU:        testCPULabelValue,
				Mem:        testMemoryLabelValue,
				Exclusive:  testExclusiveLabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeLow),
									string(entity.NodeLabelImportanceTypeMedium),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testCPULabelValue)},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testMemoryLabelValue)},
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=high,cpu,mem,exclusive",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeHigh,
				CPU:        testCPULabelValue,
				Mem:        testMemoryLabelValue,
				Exclusive:  testExclusiveLabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeLow),
									string(entity.NodeLabelImportanceTypeMedium),
									string(entity.NodeLabelImportanceTypeHigh),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testCPULabelValue)},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testMemoryLabelValue)},
							},
						},
					},
				},
			},
		},
		{
			tag: "importance=special,cpu,mem,exclusive",
			label: entity.NodeAffinityLabelConfig{
				Importance: entity.NodeLabelImportanceTypeSpecial,
				CPU:        testCPULabelValue,
				Mem:        testMemoryLabelValue,
				Exclusive:  testExclusiveLabelValue,
			},
			baseTemplateCfg: entity.NodeAffinityTemplate{
				RequiredDuringSchedulingIgnoredDuringExecution: []entity.NodeSelectorTerms{
					{
						MatchExpressions: []entity.AffinityMatchExpression{
							{
								Key:      entity.AffinityMatchExpressionKeySpec,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values: []string{
									string(entity.NodeLabelImportanceTypeSpecial),
								},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyCPU,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testCPULabelValue)},
							},
							{
								Key:      entity.AffinityMatchExpressionKeyMemory,
								Operator: entity.AffinityMatchExpressionOperatorIn,
								Values:   []string{string(testMemoryLabelValue)},
							},
						},
					},
				},
			},
		},
	}
}

// TODO: 增加云端镜像仓库区分测试
func TestService_RenderTemplate(t *testing.T) {
	nodeSelectorTestcases := generateNodeSelectorTestcases()
	labelTestcases := generateLabelTestcases()

	testLabelAndNodeSelector(t, labelTestcases, nodeSelectorTestcases,
		entity.AppTypeService, testRestfulServiceStartTask)

	testLabelAndNodeSelector(t, labelTestcases, nodeSelectorTestcases,
		entity.AppTypeService, testGRPCServiceStartTask)

	testLabelAndNodeSelector(t, labelTestcases, nodeSelectorTestcases,
		entity.AppTypeCronJob, testCronJobStartTask)

	testLabelAndNodeSelector(t, labelTestcases, nodeSelectorTestcases,
		entity.AppTypeOneTimeJob, testJobStartTask)

	for _, clusterSet := range s.k8sClusters {
		for clusterName := range clusterSet {
			ensuredClusterName := clusterName
			t.Run(string(ensuredClusterName)+"::http-service", func(t *testing.T) {
				tpl, err := s.initServiceTemplate(
					context.Background(),
					testProject,
					testRestfulServiceApp,
					testRestfulServiceStartTask,
					testTeam,
					testRestfulServiceApp.ServiceName,
				)
				assert.NoError(t, err)
				data, err := s.RenderK8sTemplate(context.Background(),
					testTemplateFileDir, ensuredClusterName, entity.AppEnvStg, tpl)
				assert.NoError(t, err)
				fmt.Printf("%s\n", data)

				res := make(map[string]interface{})
				err = yaml.Unmarshal([]byte(data), &res)
				assert.NoError(t, err)
				fmt.Printf("%v\n", res)
			})

			t.Run(string(ensuredClusterName)+"::grpc-service", func(t *testing.T) {
				tpl, err := s.initServiceTemplate(
					context.Background(),
					testProject,
					testGRPCServiceApp,
					testGRPCServiceStartTask,
					testTeam,
					testGRPCServiceApp.ServiceName,
				)
				assert.NoError(t, err)
				data, err := s.RenderK8sTemplate(context.Background(),
					testTemplateFileDir, ensuredClusterName, entity.AppEnvStg, tpl)
				assert.NoError(t, err)
				fmt.Printf("%s\n", data)

				res := make(map[string]interface{})
				err = yaml.Unmarshal([]byte(data), &res)
				assert.NoError(t, err)
				fmt.Printf("%v\n", res)
			})

			t.Run(string(ensuredClusterName)+"::config map", func(t *testing.T) {
				configData, err := s.GetAppConfig(context.Background(), &req.GetConfigManagerFileReq{
					ProjectID:  testProject.ID,
					EnvName:    entity.AppEnvStg,
					CommitID:   "c91da6fae5653c09b84e46d5438de3be234f1057",
					IsDecrypt:  true,
					FormatType: req.ConfigManagerFormatTypeJSON,
				})
				assert.NoError(t, err)

				tpl, err := s.initAppConfigMapTemplate(
					context.Background(),
					testProject,
					testRestfulServiceApp,
					testTeam,
					testRestfulServiceStartTask,
					configData.Config.(map[string]interface{}),
				)
				require.NoError(t, err)
				data, err := s.RenderK8sTemplate(context.Background(),
					testTemplateFileDir, ensuredClusterName, entity.AppEnvStg, tpl)
				assert.NoError(t, err)
				fmt.Printf("%s\n", data)

				res := make(map[string]interface{})
				err = yaml.Unmarshal([]byte(data), &res)
				assert.NoError(t, err)
				fmt.Printf("%v\n", res)
			})

			t.Run(string(ensuredClusterName)+"::AliyunLogConfig", func(t *testing.T) {
				tpl, err := s.initAliLogConfigTemplate(
					context.Background(),
					testProject,
					testRestfulServiceApp,
					testRestfulServiceStartTask,
					testTeam,
				)
				assert.NoError(t, err)
				data, err := s.RenderK8sTemplate(context.Background(),
					testTemplateFileDir, ensuredClusterName, entity.AppEnvStg, tpl)
				assert.NoError(t, err)
				fmt.Printf("%s\n", data)

				res := make(map[string]interface{})
				err = yaml.Unmarshal([]byte(data), &res)
				assert.NoError(t, err)
				fmt.Printf("%v\n", res)
			})

			t.Run(string(ensuredClusterName)+"::HPA", func(t *testing.T) {
				tpl, err := s.initHPATemplate(
					context.Background(),
					testProject,
					testRestfulServiceApp,
					testRestfulServiceStartTask,
					testTeam,
				)
				assert.NoError(t, err)
				data, err := s.RenderK8sTemplate(context.Background(),
					testTemplateFileDir, ensuredClusterName, entity.AppEnvStg, tpl)
				assert.NoError(t, err)
				fmt.Printf("%s\n", data)

				res := make(map[string]interface{})
				err = yaml.Unmarshal([]byte(data), &res)
				assert.NoError(t, err)
				fmt.Printf("%v\n", res)
			})
		}
	}

	t.Run("Jenkins", func(t *testing.T) {
		tpl, err := s.initJenkinsConfigTemplate(context.Background(), testProject, &req.CreateImageJobReq{})
		assert.NoError(t, err)
		data, err := s.RenderTemplate(context.Background(),
			"../template/jenkins/ImageConfig.xml", tpl)
		assert.NoError(t, err)
		fmt.Printf("%s\n", data)
	})

	t.Run("Jenkins-CI", func(t *testing.T) {
		tpl := s.initJenkinsCIConfigTemplate(context.Background(), testProject)
		data, err := s.RenderTemplate(context.Background(), "../template/jenkins/CIConfig.xml", tpl)
		assert.NoError(t, err)
		fmt.Printf("%s\n", data)
	})
}
