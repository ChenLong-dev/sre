package service

import (
	"rulai/config"
	"rulai/models/req"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"

	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/emirpasic/gods/sets/hashset"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/cm"
	"gitlab.shanhai.int/sre/library/net/httpclient"
)

// 需要特殊关注的配置中心 error code
const (
	ConfigManagerErrcodeMissingReplacementConfigFile = 9010020
	ConfigManagerErrcodeExtraReplacementConfigFIle   = 9010021
)

var (
	kafkaResourceRegexp = regexp.MustCompile(`common/kafka/[\w-]+\.yaml@basic`)
	etcdResourceRegexp  = regexp.MustCompile(`common/etcd/[\w-]+\.yaml@basic`)
	mysqlResourceRegexp = regexp.MustCompile(`common/mysql/r[\w-]+\.yaml@basic`)
	amqpResourceRegexp  = regexp.MustCompile(`common/amqp/[\w-]+\.yaml@basic`)
	mongoResourceRegexp = regexp.MustCompile(`common/mongo/dds-[\w-]+\.yaml@basic`)
	redisResourceRegexp = regexp.MustCompile(`common/redis/r-[\w-]+\.yaml@basic`)
)

// 获取应用配置
func (s *Service) GetAppConfig(ctx context.Context, getReq *req.GetConfigManagerFileReq) (*resp.GetAppConfigDetailResp, error) {
	response := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/v1/api/apps/%s/config", config.Conf.ConfigManager.Host, getReq.ProjectID)).
		QueryParams(
			httpclient.NewUrlValue().
				Add("env", string(getReq.EnvName)).
				Add("decrypt", strconv.FormatBool(getReq.IsDecrypt)).
				Add("format", string(getReq.FormatType)).
				Add("commit_id", getReq.CommitID).
				Add("project_name", getReq.ProjectName).
				Add("rename_prefix", getReq.ConfigRenamePrefix).
				Add("rename_mode", strconv.Itoa(int(getReq.ConfigRenameMode))),
		).
		Headers(
			httpclient.GetDefaultHeader().
				Add("AccessToken", config.Conf.ConfigManager.Token),
		).
		Method(http.MethodGet).
		AccessStatusCode(http.StatusOK, http.StatusInternalServerError).
		Fetch(ctx)

	err := response.Error()
	if err != nil {
		return nil, errors.Wrap(_errcode.ConfigManagerInternalError, err.Error())
	}

	if response.StatusCode == http.StatusOK && getReq.FormatType == req.ConfigManagerFormatTypeYaml {
		yamlData, e := response.Body()
		if e != nil {
			return nil, errors.Wrap(_errcode.ConfigManagerInternalError, e.Error())
		}

		return &resp.GetAppConfigDetailResp{
			CommitID: "",
			Config:   yamlData,
		}, nil
	}

	res := new(resp.GetAppConfigResp)
	err = response.DecodeJSON(res)
	if err != nil {
		return nil, errors.Wrap(_errcode.ConfigManagerInternalError, err.Error())
	}

	switch res.Code {
	case 0:

	case ConfigManagerErrcodeMissingReplacementConfigFile:
		return nil, errors.WithStack(_errcode.MissingReplacementConfigFile)

	case ConfigManagerErrcodeExtraReplacementConfigFIle:
		return nil, errors.WithStack(_errcode.ExtraReplacementConfigFIle)

	default:
		return nil, errors.Wrapf(_errcode.ConfigManagerInternalError, "%d:%s", res.Code, res.Message)
	}

	return res.Data, nil
}

// 从配置中心获取项目资源
func (s *Service) GetProjectResourceFromConfig(ctx context.Context, getReq *req.GetProjectResourceFromConfigReq,
	projectDetail *resp.ProjectDetailResp) (*resp.GetProjectResourceFromConfigResp, error) {
	res := new(resp.GetProjectResourceFromConfigResp)

	file, err := cm.DefaultClient().
		GetOriginFile(
			fmt.Sprintf("projects/%s/%s/config.yaml", projectDetail.Name, getReq.EnvName),
			getReq.CommitID,
		)
	if err != nil {
		// 不存在则直接返回
		if strings.Contains(err.Error(), "9000010") {
			return res, nil
		}

		return nil, errors.Wrap(_errcode.ConfigManagerInternalError, err.Error())
	}

	// kafka
	kafkaLine := kafkaResourceRegexp.FindAllString(string(file), -1)
	kafkaSet := hashset.New()

	for _, line := range kafkaLine {
		curIP := strings.ReplaceAll(
			strings.TrimPrefix(
				strings.TrimSuffix(line, ".yaml@basic"),
				"common/kafka/",
			),
			"-", ".",
		)
		kafkaSet.Add(curIP)
	}

	kafkaList := make([]*resp.AliProjectResourceResp, 0)
	for _, ip := range kafkaSet.Values() {
		kafkaList = append(kafkaList, &resp.AliProjectResourceResp{
			ID: ip.(string),
		})
	}

	res.Kafka = kafkaList

	// etcd
	etcdLine := etcdResourceRegexp.FindAllString(string(file), -1)
	etcdSet := hashset.New()

	for _, line := range etcdLine {
		curIP := strings.ReplaceAll(
			strings.TrimPrefix(
				strings.TrimSuffix(line, ".yaml@basic"),
				"common/etcd/",
			),
			"-", ".",
		)
		etcdSet.Add(curIP)
	}

	etcdList := make([]*resp.AliProjectResourceResp, 0)
	for _, ip := range etcdSet.Values() {
		etcdList = append(etcdList, &resp.AliProjectResourceResp{
			ID: ip.(string),
		})
	}

	res.Etcd = etcdList

	// amqp
	amqpLine := amqpResourceRegexp.FindAllString(string(file), -1)
	amqpSet := hashset.New()

	for _, line := range amqpLine {
		curIP := strings.ReplaceAll(
			strings.TrimPrefix(
				strings.TrimSuffix(line, ".yaml@basic"),
				"common/amqp/",
			),
			"-", ".",
		)
		amqpSet.Add(curIP)
	}

	amqpList := make([]*resp.AliProjectResourceResp, 0)
	for _, ip := range amqpSet.Values() {
		amqpList = append(amqpList, &resp.AliProjectResourceResp{
			ID: ip.(string),
		})
	}

	res.AMQP = amqpList

	// mysql
	mysqlLine := mysqlResourceRegexp.FindAllString(string(file), -1)

	mysqlSet := hashset.New()

	for _, line := range mysqlLine {
		curID := strings.TrimPrefix(
			strings.TrimSuffix(line, ".yaml@basic"),
			"common/mysql/",
		)
		mysqlSet.Add(curID)
	}

	mysqlList := make([]*resp.AliProjectResourceResp, 0)
	for _, id := range mysqlSet.Values() {
		mysqlList = append(mysqlList, &resp.AliProjectResourceResp{
			ID:         id.(string),
			ConsoleURL: fmt.Sprintf("https://rdsnext.console.aliyun.com/#/detail/%s/basicInfo", id),
		})
	}

	res.Mysql = mysqlList

	// mongo
	mongoLine := mongoResourceRegexp.FindAllString(string(file), -1)

	mongoSet := hashset.New()

	for _, line := range mongoLine {
		curID := strings.TrimPrefix(
			strings.TrimSuffix(line, ".yaml@basic"),
			"common/mongo/",
		)
		mongoSet.Add(curID)
	}

	mongoList := make([]*resp.AliProjectResourceResp, 0)

	for _, id := range mongoSet.Values() {
		mongoList = append(mongoList, &resp.AliProjectResourceResp{
			ID:         id.(string),
			ConsoleURL: fmt.Sprintf("https://next.console.aliyun.com/replicate/cn-shanghai/instances/%s/basicInfo", id),
		})
	}

	res.Mongo = mongoList

	// redis
	redisLine := redisResourceRegexp.FindAllString(string(file), -1)

	redisSet := hashset.New()

	for _, line := range redisLine {
		curID := strings.TrimPrefix(
			strings.TrimSuffix(line, ".yaml@basic"),
			"common/redis/",
		)
		redisSet.Add(curID)
	}

	redisList := make([]*resp.AliProjectResourceResp, 0)

	for _, id := range redisSet.Values() {
		redisList = append(redisList, &resp.AliProjectResourceResp{
			ID:         id.(string),
			ConsoleURL: fmt.Sprintf("https://kvstore.console.aliyun.com/#/detail/%s/Normal/VPC/info", id),
		})
	}

	res.Redis = redisList

	return res, nil
}
