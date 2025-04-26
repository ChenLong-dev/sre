package service

import (
	"rulai/config"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"

	"context"
	"fmt"
	"net/http"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/null"
	"gitlab.shanhai.int/sre/library/log"
)

// GetLogStoreDetail 获取阿里云日志仓库详情
func (s *Service) GetLogStoreDetail(ctx context.Context, getReq *req.AliGetLogStoreDetailReq) (*resp.AliGetLogStoreDetailResp, error) {
	store, err := s.GetAliLogStore(ctx, getReq)
	if err != nil {
		return nil, err
	}

	return &resp.AliGetLogStoreDetailResp{
		LogStoreName: store.Name,
		TTL:          store.TTL,
	}, nil
}

// GetAliLogStore get the whole logStore
func (s *Service) GetAliLogStore(ctx context.Context, getReq *req.AliGetLogStoreDetailReq) (*sls.LogStore, error) {
	store, err := s.aliLogClient.GetLogStore(getReq.ProjectName, getReq.StoreName)
	if err != nil {
		if aliErr, ok := err.(*sls.Error); ok && (aliErr.Code == sls.LOGSTORE_NOT_EXIST || aliErr.Code == sls.PROJECT_NOT_EXIST) {
			log.Errorc(ctx, "get ali logStore err: %+v", aliErr)
			return nil, errors.Wrap(_errcode.AliResourceNotFoundError, err.Error())
		}

		log.Errorc(ctx, "get ali logStore err: %+v", err)
		return nil, errors.Wrap(_errcode.AliSDKInternalError, err.Error())
	}

	return store, nil
}

// DeleteAliLogStore 删除阿里云日志仓库
func (s *Service) DeleteAliLogStore(ctx context.Context, delReq *req.AliDeleteLogStoreReq) error {
	err := s.aliLogClient.DeleteLogStore(delReq.ProjectName, delReq.StoreName)
	if err != nil {
		if aliErr, ok := err.(*sls.Error); ok && (aliErr.Code == sls.LOGSTORE_NOT_EXIST || aliErr.Code == sls.PROJECT_NOT_EXIST) {
			return errors.Wrap(_errcode.AliResourceNotFoundError, err.Error())
		}

		return errors.Wrap(_errcode.AliSDKInternalError, err.Error())
	}

	return nil
}

// GetLogStoreIndex 获取logstore索引
func (s *Service) GetLogStoreIndex(ctx context.Context, getReq *req.AliGetLogStoreIndexReq) (*sls.Index, error) {
	index, err := s.aliLogClient.GetIndex(getReq.ProjectName, getReq.StoreName)
	if err != nil {
		if aliErr, ok := err.(*sls.Error); ok && (aliErr.Code == sls.LOGSTORE_NOT_EXIST || aliErr.Code == sls.PROJECT_NOT_EXIST) {
			return nil, errors.Wrap(_errcode.AliResourceNotFoundError, err.Error())
		}

		return nil, errors.Wrap(_errcode.AliSDKInternalError, err.Error())
	}

	return index, err
}

// AddLogStoreIndex 添加logstore索引
func (s *Service) AddLogStoreIndex(ctx context.Context, addReq *req.AliAddLogStoreIndexesReq) error {
	index, err := s.GetLogStoreIndex(ctx, &req.AliGetLogStoreIndexReq{
		ProjectName: addReq.ProjectName,
		StoreName:   addReq.StoreName,
	})
	if err != nil {
		return err
	}

	hasIndex := true
	for name, key := range addReq.Index {
		// 过滤已添加索引
		if _, ok := index.Keys[name]; ok {
			continue
		}

		tokens := key.Token
		if key.Type == entity.AliLogIndexTypeText && len(tokens) == 0 {
			tokens = entity.DefaultAliLogIndexToken
		}
		key.Token = tokens

		index.Keys[name] = key
		hasIndex = false
	}

	// 若所有索引已存在，则不再更新
	if hasIndex {
		return nil
	}

	err = s.UpdateAliLogIndex(ctx, &req.AliUpdateStoreLogIndexReq{
		ProjectName: addReq.ProjectName,
		StoreName:   addReq.StoreName,
		Index:       *index,
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdateAliLogIndex 更新logstore索引
func (s *Service) UpdateAliLogIndex(ctx context.Context, updateReq *req.AliUpdateStoreLogIndexReq) error {
	err := s.aliLogClient.UpdateIndex(updateReq.ProjectName, updateReq.StoreName, updateReq.Index)
	if err != nil {
		if aliErr, ok := err.(*sls.Error); ok && (aliErr.Code == sls.LOGSTORE_NOT_EXIST || aliErr.Code == sls.PROJECT_NOT_EXIST) {
			return errors.Wrap(_errcode.AliResourceNotFoundError, err.Error())
		}

		return errors.Wrap(_errcode.AliSDKInternalError, err.Error())
	}

	return nil
}

// CreateAliLogIndexFromLogConfig creates index for logstore and checks the result
func (s *Service) CreateAliLogIndexFromLogConfig(ctx context.Context,
	clusterName entity.ClusterName, createReq *req.AliCreateStoreLogIndexReq) error {
	aliconfig, err := s.GetAliLogConfigDetail(ctx, clusterName,
		&req.GetAliLogConfigDetailReq{
			Namespace: createReq.Namespace,
			Name:      createReq.AliLogConfigName,
			Env:       createReq.EnvName,
		})
	if err != nil {
		return err
	}

	// 获取logstore名称
	logStoreName, ok, err := unstructured.NestedString(aliconfig.UnstructuredContent(), "spec", "logstore")
	if err != nil {
		return err
	}
	if !ok || (ok && logStoreName == "") {
		return errors.Wrap(_errcode.K8sResourceNotFoundError, "logstore not found")
	}

	// 获取集群名
	// 获取日志服务项目名
	projectName := config.Conf.Other.AliLogProjectStgName
	if createReq.Namespace == string(entity.AppEnvPre) || createReq.Namespace == string(entity.AppEnvPrd) {
		projectName = config.Conf.Other.AliLogProjectPrdName
	}

	// 添加索引
	indices := make(map[string]sls.IndexKey)
	for _, index := range entity.LogStoreIndexKeys {
		indices[string(index)] = sls.IndexKey{
			Type:     entity.AliLogIndexTypeText,
			DocValue: true,
		}
	}
	err = s.AddLogStoreIndex(ctx, &req.AliAddLogStoreIndexesReq{
		ProjectName: projectName,
		StoreName:   logStoreName,
		Index:       indices,
	})
	if err != nil {
		return err
	}

	// check index creation
	// 获取索引
	index, err := s.GetLogStoreIndex(ctx, &req.AliGetLogStoreIndexReq{
		ProjectName: projectName,
		StoreName:   logStoreName,
	})
	if err != nil {
		return err
	}

	// 检索索引创建
	for _, key := range entity.LogStoreIndexKeys {
		_, ok := index.Keys[string(key)]
		if !ok {
			return errors.Wrap(_errcode.LogStoreIndexNotFoundError, string(key))
		}
	}

	return nil
}

// getAliLogStoreColdStorageShipperName return shipper name
func (s *Service) getAliLogStoreColdStorageShipperName(logStoreName string) string {
	return fmt.Sprintf("%s-cold-shipper", logStoreName)
}

// createAliLogStoreShipperWhenNotExist create logStore to oss when shipper does not exist
func (s *Service) createAliLogStoreShipperWhenNotExist(ctx context.Context,
	createReq *req.CreateAliLogStoreShipperReq) error {
	logStore, err := s.GetAliLogStore(ctx, &req.AliGetLogStoreDetailReq{
		ProjectName: createReq.LogStoreProject,
		StoreName:   createReq.LogStoreName,
	})
	if err != nil {
		return err
	}

	_, err = logStore.GetShipper(createReq.ShipperName)
	// 存在时跳过创建阶段
	if err == nil {
		return nil
	}

	if s.isAliLogStoreShipperNotExistError(err) {
		return logStore.CreateShipper(&sls.Shipper{
			ShipperName:         createReq.CreateOSSShipperReq.ShipperName,
			TargetType:          createReq.TargetType,
			TargetConfiguration: createReq.CreateOSSShipperReq,
		})
	}

	log.Errorc(ctx, "get shipper err occurred when create shipper, err: %s", err)
	return err
}

// deleteAliLogStoreShipperWhenExist delete logStore shipper when exist
func (s *Service) deleteAliLogStoreShipperWhenExist(ctx context.Context, delReq *req.DeleteAliLogStoreShipperReq) error {
	logStore, err := s.GetAliLogStore(ctx, &req.AliGetLogStoreDetailReq{
		ProjectName: delReq.LogStoreProject,
		StoreName:   delReq.LogStoreName,
	})
	if err != nil {
		return err
	}

	_, err = logStore.GetShipper(delReq.ShipperName)
	if err != nil {
		// 不存在时跳过删除阶段
		if s.isAliLogStoreShipperNotExistError(err) {
			return nil
		}

		log.Errorc(ctx, "get shipper error occurred when delete shipper, err: %s", err)
		return err
	}

	// Delete cold storage shipper
	return logStore.DeleteShipper(delReq.ShipperName)
}

// Sync log store cold storage shipper
// create shipper when open cold storage, delete shipper when close cold storage
func (s *Service) SyncAliLogStoreColdStorageShipper(ctx context.Context,
	clusterName entity.ClusterName, syncReq *req.SyncAliLogStoreColdStorageShipperReq) error {
	logConfigList, err := s.ListAliLogConfig(ctx, clusterName, string(syncReq.EnvName),
		&req.ListAliLogConfigReq{
			// todo	istio 这里的 env 换成 ns 是不是会影响日志绑定
			Namespace:       string(syncReq.EnvName),
			ProjectName:     syncReq.ProjectName,
			OpenColdStorage: null.BoolFrom(true),
		})
	if err != nil {
		return err
	}

	logStoreProject, ossBucket := config.Conf.Other.AliLogProjectStgName, config.Conf.ColdStorage.OSSStgBucket
	if syncReq.EnvName == entity.AppEnvPrd || syncReq.EnvName == entity.AppEnvPre {
		logStoreProject = config.Conf.Other.AliLogProjectPrdName
		ossBucket = config.Conf.ColdStorage.OSSPrdBucket
	}
	// Create logStore shipper to oss when not exist
	if len(logConfigList.Items) > 0 {
		if err := s.createAliLogStoreShipperWhenNotExist(ctx, &req.CreateAliLogStoreShipperReq{
			TargetType:      sls.OSSShipperType,
			ProjectName:     syncReq.ProjectName,
			LogStoreName:    syncReq.LogStoreName,
			LogStoreProject: logStoreProject,
			ShipperName:     s.getAliLogStoreColdStorageShipperName(syncReq.LogStoreName),
			CreateOSSShipperReq: &req.CreateAliLogStoreOSSShipperReq{
				ShipperName:    s.getAliLogStoreColdStorageShipperName(syncReq.LogStoreName),
				OSSBucket:      ossBucket,
				OSSPrefix:      syncReq.ProjectName,
				RoleArn:        config.Conf.ColdStorage.RoleArn,
				BufferInterval: config.Conf.ColdStorage.BufferInterval,
				BufferSize:     config.Conf.ColdStorage.BufferSize,
				CompressType:   config.Conf.ColdStorage.CompressType,
				PathFormat:     config.Conf.ColdStorage.PathFormat,
				Format:         config.Conf.ColdStorage.Format,
			},
		}); err != nil {
			return err
		}
		return nil
	}

	// Delete cold storage shipper when exist
	if err := s.deleteAliLogStoreShipperWhenExist(ctx, &req.DeleteAliLogStoreShipperReq{
		LogStoreProject: logStoreProject,
		LogStoreName:    syncReq.LogStoreName,
		ShipperName:     s.getAliLogStoreColdStorageShipperName(syncReq.LogStoreName),
	}); err != nil {
		return err
	}

	return nil
}

// isAliLogStoreShipperNotExistError 检验错误是否是日志冷存不存在的特殊错误
// 阿里云目前可能会返回两种该情况的响应, 需要兼容
//  1. HTTPCode=400, Code=sls.PARAMETER_INVALID)
//  2. HTTPCode=404, Code=sls.SHIPPER_NOT_EXIST)
func (s *Service) isAliLogStoreShipperNotExistError(err error) bool {
	if err == nil {
		return false
	}

	aliErr, ok := err.(*sls.Error)
	return ok && (aliErr.HTTPCode == http.StatusBadRequest && aliErr.Code == sls.PARAMETER_INVALID) ||
		(aliErr.HTTPCode == http.StatusNotFound && aliErr.Code == sls.SHIPPER_NOT_EXIST)
}
