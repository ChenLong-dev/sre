package service

import (
	"rulai/config"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/utils"
	_errcode "rulai/utils/errcode"

	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"gitlab.shanhai.int/sre/library/net/httpclient"
)

const (
	PrometheusDefaultMaxTotalCPUCountTime          = "1d"
	PrometheusDefaultMaxTotalMemCountTime          = "1d"
	PrometheusDefaultWasteMaxCPUUsageRateCountTime = "1d"
	PrometheusDefaultWasteMaxMemUsageRateCountTime = "1d"
)

// 解析prometheus数据
func (s *Service) renderAndGetPrometheusData(ctx context.Context, envName entity.AppEnvName,
	tplPath string, tplReq interface{}) (float64, error) {
	tpl, err := s.RenderTemplate(ctx, tplPath, tplReq)
	if err != nil {
		return 0, err
	}

	data, err := s.GetPrometheusData(ctx, &req.QueryPrometheusReq{
		EnvName: envName,
		SQL:     tpl,
	})
	if err != nil {
		return 0, err
	}

	// 查不到
	if len(data.Result) == 0 {
		return 0, _errcode.PrometheusQueryEmptyError
	}
	// 格式错误
	if len(data.Result[0].Value) != 2 {
		return 0, _errcode.PrometheusInternalError
	}
	resString, ok := data.Result[0].Value[1].(string)
	if !ok {
		return 0, _errcode.PrometheusInternalError
	}
	// 转换
	res, err := strconv.ParseFloat(resString, 64)
	if err != nil {
		return 0, errors.Wrap(errcode.InternalError, err.Error())
	}

	return res, nil
}

// 获取最大的总cpu时间
func (s *Service) GetMaxTotalCPUTime(ctx context.Context, queryReq *req.GetMaxTotalCPUReq,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp) (float64, error) {
	if queryReq.CountTime == "" {
		queryReq.CountTime = PrometheusDefaultMaxTotalCPUCountTime
	}
	containerName := utils.GetPodContainerName(project.Name, app.Name)
	labelName := config.Conf.Prometheus.StgContainerLabelName
	if queryReq.EnvName == entity.AppEnvPrd || queryReq.EnvName == entity.AppEnvPre {
		labelName = config.Conf.Prometheus.PrdContainerLabelName
	}

	res, err := s.renderAndGetPrometheusData(ctx, queryReq.EnvName,
		"./template/prometheus/sql/MaxTotalCpuTime.sql",
		&entity.MaxTotalCPUTemplate{
			EnvName:            queryReq.EnvName,
			ContainerName:      containerName,
			CountTime:          queryReq.CountTime,
			ContainerLabelName: labelName,
		},
	)
	if err != nil {
		return 0, err
	}
	return res, nil
}

// 获取最小的总cpu时间
func (s *Service) GetMinTotalCPUTime(ctx context.Context, queryReq *req.GetMinTotalCPUReq,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp) (float64, error) {
	if queryReq.CountTime == "" {
		queryReq.CountTime = PrometheusDefaultMaxTotalCPUCountTime
	}
	containerName := utils.GetPodContainerName(project.Name, app.Name)
	labelName := config.Conf.Prometheus.StgContainerLabelName
	if queryReq.EnvName == entity.AppEnvPrd || queryReq.EnvName == entity.AppEnvPre {
		labelName = config.Conf.Prometheus.PrdContainerLabelName
	}

	res, err := s.renderAndGetPrometheusData(ctx, queryReq.EnvName,
		"./template/prometheus/sql/MinTotalCpuTime.sql",
		&entity.MaxTotalCPUTemplate{
			EnvName:            queryReq.EnvName,
			ContainerName:      containerName,
			CountTime:          queryReq.CountTime,
			ContainerLabelName: labelName,
		},
	)
	if err != nil {
		return 0, err
	}
	return res, nil
}

// 获取最大的总内存大小
func (s *Service) GetMaxTotalMemBytes(ctx context.Context, queryReq *req.GetMaxTotalMemReq,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp) (float64, error) {
	if queryReq.CountTime == "" {
		queryReq.CountTime = PrometheusDefaultMaxTotalMemCountTime
	}
	containerName := utils.GetPodContainerName(project.Name, app.Name)
	labelName := config.Conf.Prometheus.StgContainerLabelName
	if queryReq.EnvName == entity.AppEnvPrd || queryReq.EnvName == entity.AppEnvPre {
		labelName = config.Conf.Prometheus.PrdContainerLabelName
	}

	res, err := s.renderAndGetPrometheusData(ctx, queryReq.EnvName,
		"./template/prometheus/sql/MaxTotalMemBytes.sql",
		&entity.MaxTotalMemTemplate{
			EnvName:            queryReq.EnvName,
			ContainerName:      containerName,
			CountTime:          queryReq.CountTime,
			ContainerLabelName: labelName,
		},
	)
	if err != nil {
		return 0, err
	}
	return res, nil
}

// 获取最大的总内存大小
func (s *Service) GetMinTotalMemBytes(ctx context.Context, queryReq *req.GetMinTotalMemReq,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp) (float64, error) {
	if queryReq.CountTime == "" {
		queryReq.CountTime = PrometheusDefaultMaxTotalMemCountTime
	}
	containerName := utils.GetPodContainerName(project.Name, app.Name)
	labelName := config.Conf.Prometheus.StgContainerLabelName
	if queryReq.EnvName == entity.AppEnvPrd || queryReq.EnvName == entity.AppEnvPre {
		labelName = config.Conf.Prometheus.PrdContainerLabelName
	}

	res, err := s.renderAndGetPrometheusData(ctx, queryReq.EnvName,
		"./template/prometheus/sql/MinTotalMemBytes.sql",
		&entity.MaxTotalMemTemplate{
			EnvName:            queryReq.EnvName,
			ContainerName:      containerName,
			CountTime:          queryReq.CountTime,
			ContainerLabelName: labelName,
		},
	)
	if err != nil {
		return 0, err
	}
	return res, nil
}

// 获取浪费的最大cpu使用率
// 若值为0，则认为不浪费
func (s *Service) GetWastedMaxCPUUsageRate(ctx context.Context, queryReq *req.GetWastedMaxCPUUsageRateReq,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp) (float64, error) {
	if queryReq.CountTime == "" {
		queryReq.CountTime = PrometheusDefaultWasteMaxCPUUsageRateCountTime
	}
	containerName := utils.GetPodContainerName(project.Name, app.Name)
	labelName := config.Conf.Prometheus.StgContainerLabelName
	if queryReq.EnvName == entity.AppEnvPrd || queryReq.EnvName == entity.AppEnvPre {
		labelName = config.Conf.Prometheus.PrdContainerLabelName
	}

	res, err := s.renderAndGetPrometheusData(ctx, queryReq.EnvName,
		"./template/prometheus/sql/WastedMaxCpuUsageRate.sql",
		&entity.WastedMaxCPUUsageRateTemplate{
			EnvName:            queryReq.EnvName,
			ContainerLabelName: labelName,
			ContainerName:      containerName,
			CountTime:          queryReq.CountTime,
			// 20%
			UsageRateLimit: 0.2,
			MinCPUResource: entity.CPUResourceNano,
		},
	)
	// 查不到则认为不浪费
	if errcode.EqualError(_errcode.PrometheusQueryEmptyError, err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return res, nil
}

// 获取浪费的最大内存使用率
// 若值为0，则认为不浪费
func (s *Service) GetWastedMaxMemUsageRate(ctx context.Context, queryReq *req.GetWastedMaxMemUsageRateReq,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp) (float64, error) {
	if queryReq.CountTime == "" {
		queryReq.CountTime = PrometheusDefaultWasteMaxMemUsageRateCountTime
	}
	containerName := utils.GetPodContainerName(project.Name, app.Name)
	labelName := config.Conf.Prometheus.StgContainerLabelName
	if queryReq.EnvName == entity.AppEnvPrd || queryReq.EnvName == entity.AppEnvPre {
		labelName = config.Conf.Prometheus.PrdContainerLabelName
	}

	res, err := s.renderAndGetPrometheusData(ctx, queryReq.EnvName,
		"./template/prometheus/sql/WastedMaxMemUsageRate.sql",
		&entity.WastedMaxMemUsageRateTemplate{
			EnvName:            queryReq.EnvName,
			ContainerLabelName: labelName,
			ContainerName:      containerName,
			CountTime:          queryReq.CountTime,
			// 20%
			UsageRateLimit: 0.2,
			MinMemResource: entity.MemResourceNanoBytes,
		},
	)
	// 查不到则认为不浪费
	if errcode.EqualError(_errcode.PrometheusQueryEmptyError, err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s *Service) GetPrometheusData(ctx context.Context, queryReq *req.QueryPrometheusReq) (*resp.QueryPrometheusDataResp, error) {
	host := config.Conf.Prometheus.StgHost
	if queryReq.EnvName == entity.AppEnvPrd || queryReq.EnvName == entity.AppEnvPre {
		host = config.Conf.Prometheus.PrdHost
	}

	res := new(resp.QueryPrometheusResp)
	err := s.httpClient.Builder().
		Method(http.MethodGet).
		URL(fmt.Sprintf("%s/api/v1/query", host)).
		QueryParams(httpclient.NewUrlValue().Add("query", queryReq.SQL)).
		Headers(httpclient.GetDefaultHeader()).
		DisableBreaker(true).
		Fetch(ctx).
		DecodeJSON(res)
	if err != nil {
		return nil, errors.Wrap(_errcode.PrometheusInternalError, err.Error())
	} else if res.Status != resp.PrometheusStatusSuccess {
		return nil, errors.Wrapf(_errcode.PrometheusInternalError, "type:%s error:%s", res.ErrorType, res.Error)
	}

	return res.Data, nil
}
