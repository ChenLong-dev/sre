package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.shanhai.int/sre/library/net/errcode"

	"rulai/models/entity"
)

func TestService_CheckProjectLabelsLegal(t *testing.T) {
	t.Run("k8sSuccessCases", func(t *testing.T) {
		k8sSuccessCaseWithoutImportanceLabel := []string{
			"simple",
			"now-with-dashes",
			"1-starts-with-num",
			"1234",
			"3-num",
			"UpperCaseAreOK123",
		}
		assert.True(t, errcode.EqualError(errcode.InvalidParams,
			s.CheckProjectLabelsLegal(k8sSuccessCaseWithoutImportanceLabel)))

		importanceLabels := []entity.ProjectLabelValue{
			entity.ProjectLabelP0,
			entity.ProjectLabelP1,
			entity.ProjectLabelP2,
		}
		for _, l := range importanceLabels {
			assert.NoError(t, s.CheckProjectLabelsLegal(append(k8sSuccessCaseWithoutImportanceLabel, string(l))))
		}
	})

	t.Run("errorNameCases", func(t *testing.T) {
		labelNameErrorCases := [][]string{
			{"nospecialchars^=@"},
			{"cantendwithadash-"},
			{"only/one/slash"},
			{strings.Repeat("a", 254)},
			{string(entity.ProjectLabelP0), string(entity.ProjectLabelP0)},
			{string(entity.ProjectLabelP0), string(entity.ProjectLabelP1)},
			{string(entity.ProjectLabelP0), string(entity.ProjectLabelP2)},
			{string(entity.ProjectLabelP1), string(entity.ProjectLabelP2)},
		}

		for i := range labelNameErrorCases {
			assert.True(t, errcode.EqualError(errcode.InvalidParams,
				s.CheckProjectLabelsLegal(labelNameErrorCases[i])))
		}
	})
}
