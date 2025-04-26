package service

import (
	"rulai/dao"
	"rulai/models/entity"
	"rulai/models/resp"

	"context"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"

	"k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func (s *Service) GetProjectLabels(ctx context.Context) ([]*resp.ProjectLabelsResp, error) {
	res := make([]*resp.ProjectLabelsResp, 0)
	labels, err := s.dao.GetProjectLabels(ctx, bson.M{}, dao.MongoFindOptionWithSortByIDAsc)
	if err != nil {
		return nil, err
	}

	err = deepcopy.Copy(&labels).To(&res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	return res, nil
}

// CheckProjectLabelsLegal 校验项目标签是否合法(有且仅有一个等级标签且通过 k8s 标签校验)
func (s *Service) CheckProjectLabelsLegal(labels []string) error {
	foundImportance := false
	labelMap := make(map[string]string)
	for _, label := range labels {
		labelMap[s.getMetricsLabelKeyName(label)] = MetricsLabelEnable
		if entity.ProjectLabelValue(label).IsProjectImportanceLevelLabel() {
			if foundImportance {
				return errors.Wrap(errcode.InvalidParams, "服务分级不支持多选")
			}
			foundImportance = true
		}
	}

	if !foundImportance {
		return errors.Wrap(errcode.InvalidParams, "必须选择一个服务分级标签(P0/P1/P2...)")
	}

	errs := validation.ValidateLabels(labelMap, field.NewPath("labels"))
	if len(errs) != 0 {
		return errors.Wrapf(errcode.InvalidParams, "label illegal err: %v", errs)
	}

	return nil
}
