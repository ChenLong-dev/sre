package huawei

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.shanhai.int/sre/library/base/null"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"gitlab.shanhai.int/sre/library/net/httpclient"

	"rulai/models/entity"
	"rulai/models/req"

	"rulai/models/resp"
	code "rulai/utils/errcode"
)

// createHuaweiLTSStream 创建华为云 LTS 日志流
// 官方文档: https://support.huaweicloud.com/api-lts/lts_api_0016.html
func (c *Controller) createHuaweiLTSStream(ctx context.Context, stream *entity.AppLTSStream, ttlInDays null.Int) (string, error) {
	clusterInfo, err := c.getClusterConfig(stream.EnvName, stream.ClusterName)
	if err != nil {
		return "", err
	}

	errResp := new(resp.HuaweiLTSStandardResp)
	createResp := new(resp.CreateHuaweiLTSStreamResp)
	createReq := &req.CreateHuaweiLTSStreamReq{LogStreamName: stream.StreamName, TTLInDays: ttlInDays}
	url := fmt.Sprintf("https://%s/v2/%s/groups/%s/streams", c.LogEndpoint, c.ProjectID, clusterInfo.LogGroupID)

	// 由于华为云目前没有提供单独查询日志流的 API, 尝试创建前无法确认日志流是否存在(有全量查询 API 但生产环境不适合)
	// 故通过调用创建 API 的返回码判断日志流是否存在, 已存在时返回特殊错误
	err = c.doHuaweiAPIRequest(ctx, http.MethodPost, url, nil, createReq, createResp, errResp, duplicateLogStreamNameWrapper)
	if err != nil {
		if errcode.EqualError(code.LTSResourceAlreadyExists, err) {
			return c.GetLTSStreamIDByStreamName(ctx, stream.EnvName, stream.ClusterName, stream.StreamName)
		}

		return "", err
	}

	return createResp.LogStreamID, nil
}

// GetLTSStreamIDByStreamName 查询指定日志组下的指定日志流名称的日志流 ID
func (c *Controller) GetLTSStreamIDByStreamName(ctx context.Context, envName entity.AppEnvName,
	clusterName entity.ClusterName, streamName string) (streamID string, err error) {
	list, err := c.ListLTSStream(ctx, envName, clusterName)
	if err != nil {
		return "", err
	}

	// 查找 logStreamName
	for _, item := range list.LogStreams {
		if item.LogStreamName == streamName {
			return item.LogStreamID, nil
		}
	}

	return "", nil
}

// ListLTSStream 查询指定日志组下的所有日志流
// 官方文档: https://support.huaweicloud.com/api-lts/lts_api_0017.html
func (c *Controller) ListLTSStream(ctx context.Context, envName entity.AppEnvName,
	clusterName entity.ClusterName) (*resp.ListLTSStreamResp, error) {
	clusterInfo, err := c.getClusterConfig(envName, clusterName)
	if err != nil {
		return nil, err
	}

	errResp := new(resp.HuaweiLTSStandardResp)
	listResp := new(resp.ListLTSStreamResp)
	listReq := &req.ListLTSStreamReq{
		ProjectID:  c.ProjectID,
		LogGroupID: clusterInfo.LogGroupID,
	}
	url := fmt.Sprintf("https://%s/v2/%s/groups/%s/streams", c.LogEndpoint, c.ProjectID, clusterInfo.LogGroupID)

	err = c.doHuaweiAPIRequest(ctx, http.MethodGet, url, nil, listReq, listResp, errResp)
	if err != nil {
		return nil, err
	}
	return listResp, nil
}

// deleteHuaweiLTSStream 删除华为云 LTS 日志流
// 官方文档: https://support.huaweicloud.com/api-lts/lts_api_0018.html
// NOTE: 删除时两种关联资源的影响有区别
//  1. 如果日志流配置了日志转储, 必须先取消日志转储才能删除日志流(这是暂时未接入日志转储的原因)
//  2. AOM 关联日志流并不影响日志流的删除
func (c *Controller) deleteHuaweiLTSStream(ctx context.Context,
	envName entity.AppEnvName, clusterName entity.ClusterName, logStreamID string) error {
	clusterInfo, err := c.getClusterConfig(envName, clusterName)
	if err != nil {
		return err
	}

	errResp := new(resp.HuaweiLTSStandardResp)
	url := fmt.Sprintf("https://%s/v2/%s/groups/%s/streams/%s", c.LogEndpoint, c.ProjectID, clusterInfo.LogGroupID, logStreamID)
	return c.doHuaweiAPIRequest(ctx, http.MethodDelete, url,
		nil, nil, nil, errResp, logStreamNotFoundWrapper, logStreamAssociatedByTransferWrapper)
}

// createHuaweiAOMToLTSRule 创建华为云 AOM 到 LTS 日志流的接入规则
// 官方文档: https://support.huaweicloud.com/api-lts/lts_api_0064.html
func (c *Controller) createHuaweiAOMToLTSRule(ctx context.Context, envName entity.AppEnvName,
	clusterName entity.ClusterName, namespace, ruleName, containerName, streamID, streamName string) (string, error) {
	clusterInfo, err := c.getClusterConfig(envName, clusterName)
	if err != nil {
		return "", err
	}

	// 不按照指定 deployments 的方式接入, 按照指定 container_name 的方式接入
	createReq := &req.CreateHuaweiAOMToLTSStreamMappingReq{
		RuleName: ruleName,
		RuleInfo: &req.HuaweiAOMToLTSStreamMappingInfo{
			ClusterID:     clusterInfo.ClusterID,
			ClusterName:   clusterInfo.ClusterNameInVendor,
			Deployments:   []string{req.HuaweiLTSDeploymentTagAll},
			ContainerName: containerName,
			Namespace:     namespace,
			Files: []*req.HuaweiAOMToLTSStreamMappingFileInfo{
				{
					FileName: req.HuaweiLTSFileNameTagAll,
					LogStreamInfo: &req.HuaweiLTSLogStreamInfo{
						TargetLogGroupID:   clusterInfo.LogGroupID,
						TargetLogGroupName: clusterInfo.LogGroupName,
						// FIXME: 华为云 API 目前有 bug, 填写日志流名称时可能会报重名错误, 当前只能先填写 target_log_stream_id
						// FIXME: ({"code":"LTS.0746","details":"AOM mapping rule log stream name already exist in another log group"})
						// FIXME: 而不填写日志流名称会关联不到 LTS
						// FIXME: 故当前只支持全部填写 target_log_stream_id 和 target_log_stream_name
						// FIXME: 待未来华为云调整后再考虑参数调整
						TargetLogStreamID:   streamID,
						TargetLogStreamName: streamName,
					},
				},
			},
		},
		ProjectID: c.ProjectID,
	}

	queryParams := httpclient.NewUrlValue()
	queryParams.Set("isBatch", "false")

	errResp := new(resp.HuaweiLTSCodeDetailsInMessageResp)
	var createResp resp.CreateHuaweiAOMToLTSStreamMappingResp
	url := fmt.Sprintf("https://%s/v2/%s/lts/aom-mapping", c.LogEndpoint, c.ProjectID)

	err = c.doHuaweiAPIRequest(ctx, http.MethodPost, url, queryParams, createReq,
		&createResp, errResp, duplicateAOMMappingRuleNameWrapper)
	if err != nil {
		if errcode.EqualError(code.LTSResourceAlreadyExists, err) {
			return c.GetAOMToLTSRuleIDByStream(ctx, envName, clusterName, namespace,
				ruleName, containerName, streamID, streamName)
		}
		return "", err
	}

	// 返回结果是数组格式, 并不是文档描述的对象格式
	if len(createResp) != 1 {
		return "", errors.Wrapf(errcode.InternalError, "unexpected aom-mapping response length(%d)", len(createResp))
	}

	return createResp[0].RuleID, nil
}

// GetAOMToLTSRuleIDByStream 根据 LTSLogStreamID 找到对应的 AOMToLTSRuleID
func (c *Controller) GetAOMToLTSRuleIDByStream(ctx context.Context, envName entity.AppEnvName,
	clusterName entity.ClusterName, namespace, ruleName, containerName, streamID, streamName string) (string, error) {
	list, err := c.ListAOMToLTSRule(ctx, envName, clusterName, namespace, ruleName, containerName, streamID, streamName)
	if err != nil {
		return "", err
	}

	for _, item := range list {
		for _, streamInfo := range item.RuleInfo.Files {
			if streamInfo.LogStreamInfo.TargetLogStreamID == streamID {
				return item.RuleID, nil
			}
		}
	}

	return "", nil
}

// ListAOMToLTSRule 查询所有接入规则
// 官方文档: https://support.huaweicloud.com/api-lts/lts_api_0067.html
func (c *Controller) ListAOMToLTSRule(ctx context.Context, envName entity.AppEnvName, clusterName entity.ClusterName,
	namespace, ruleName, containerName, streamID, streamName string) (resp.ListAOMToLTSRuleResp, error) {
	clusterInfo, err := c.getClusterConfig(envName, clusterName)
	if err != nil {
		return nil, err
	}

	errResp := new(resp.HuaweiLTSCodeDetailsInMessageResp)
	var listResp resp.ListAOMToLTSRuleResp
	url := fmt.Sprintf("https://%s/v2/%s/lts/aom-mapping", c.LogEndpoint, c.ProjectID)

	query := httpclient.NewUrlValue()
	query.Add("log_group_name", clusterInfo.LogGroupName)
	query.Add("log_stream_name", streamName)

	err = c.doHuaweiAPIRequest(ctx, http.MethodGet, url, query, nil,
		&listResp, errResp, duplicateAOMMappingRuleNameWrapper)
	if err != nil {
		return nil, err
	}

	return listResp, nil
}

// deleteHuaweiAOMToLTSRule 删除华为云 AOM 到 LTS 日志流的接入规则
// FIXME: 官方文档有问题, id 参数通过 query 传递而不是通过 body; 且实际响应是个字符串数组, 当前不解析这种不规范的响应
// NOTE: 该 API 从实际效果来看, 应该是一个批量删除 API, AMS 目前只需要实现单个删除的功能
// 官方文档: https://support.huaweicloud.com/api-lts/lts_api_0066.html
func (c *Controller) deleteHuaweiAOMToLTSRule(ctx context.Context, ruleID string) error {
	queryParams := httpclient.NewUrlValue()
	queryParams.Set("id", ruleID)
	errResp := new(resp.HuaweiLTSCodeDetailsInMessageResp)
	url := fmt.Sprintf("https://%s/v2/%s/lts/aom-mapping", c.LogEndpoint, c.ProjectID)

	return c.doHuaweiAPIRequest(ctx, http.MethodDelete, url, queryParams, nil,
		nil, errResp, invalidAOMMappingRuleIDWrapper)
}

func toString(errResp resp.HuaweiLTSResp) string {
	return fmt.Sprintf("request_id(%s), error_code(%s), error_msg(%s)",
		errResp.GetRequestID(), errResp.GetErrorCode(), errResp.GetErrorMsg())
}
