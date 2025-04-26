package service

// 日志采集器相关逻辑
// TODO: 未来期望全部接入至运营商层面
import (
	"context"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"

	"rulai/config"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"
)

// ApplyLogConfig 声明式创建/更新日志采集器配置
func (s *Service) ApplyLogConfig(ctx context.Context,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp, task *resp.TaskDetailResp, team *resp.TeamDetailResp) error {
	clusterInfo, err := s.getClusterInfo(task.ClusterName, string(task.EnvName))
	if err != nil {
		return err
	}

	if clusterInfo.disableLogConfig {
		// 禁用日志采集器时返回特殊错误供判断
		// TODO: 阿里日志采集器逻辑迁移到 controller 中后, disableLogConfig 应当移至 controller 层面
		return _errcode.LogConfigDisabled
	}

	switch clusterInfo.vendor {
	case entity.VendorAli:
		// 渲染模版
		data, e := s.RenderAliLogConfigTemplate(ctx, project, app, task, team)
		if e != nil {
			return e
		}
		// 生成AliLogConfig
		_, e = s.ApplyAliLogConfig(ctx, task.ClusterName, string(task.EnvName), data)
		return e

	case entity.VendorHuawei:
		c, e := s.getVendorController(clusterInfo.vendor)
		if e != nil {
			return e
		}

		return c.ApplyLogConfig(ctx, project, app, task, team)

	default:
	}

	return errors.Wrapf(_errcode.UnknownVendorError,
		"cluster(%s), env(%s), vendor(%s)", clusterInfo.name, clusterInfo.envName, clusterInfo.vendor)
}

// LogConfigExistanceCheck 检验日志采集器配置是否存在(某些云服务商[比如阿里云]的日志配置并非创建完毕就立即存在)
func (s *Service) LogConfigExistanceCheck(ctx context.Context, app *resp.AppDetailResp, task *resp.TaskDetailResp) error {
	clusterInfo, err := s.getClusterInfo(task.ClusterName, string(task.EnvName))
	if err != nil {
		return err
	}

	switch clusterInfo.vendor {
	case entity.VendorAli:
		_, e := s.GetAliLogConfigDetail(ctx, task.ClusterName, &req.GetAliLogConfigDetailReq{
			// todo 阿里目前不支持 istio
			Namespace: task.Namespace,
			Name:      app.AliLogConfigName,
			Env:       string(task.EnvName),
		})
		return e

	case entity.VendorHuawei:
		c, e := s.getVendorController(clusterInfo.vendor)
		if e != nil {
			return e
		}

		return c.LogConfigExistanceCheck(ctx, task.ClusterName, task.EnvName, app)

	default:
	}

	return errors.Wrapf(_errcode.UnknownVendorError,
		"cluster(%s), env(%s), vendor(%s)", clusterInfo.name, clusterInfo.envName, clusterInfo.vendor)
}

// EnsureLogIndex 创建日志索引
func (s *Service) EnsureLogIndex(ctx context.Context, app *resp.AppDetailResp, task *resp.TaskDetailResp) error {
	clusterInfo, err := s.getClusterInfo(task.ClusterName, string(task.EnvName))
	if err != nil {
		return err
	}

	switch clusterInfo.vendor {
	case entity.VendorAli:
		return s.CreateAliLogIndexFromLogConfig(ctx, task.ClusterName, &req.AliCreateStoreLogIndexReq{
			Namespace:        string(task.EnvName),
			AliLogConfigName: app.AliLogConfigName,
			EnvName:          string(task.EnvName),
		})

	case entity.VendorHuawei:
		c, e := s.getVendorController(clusterInfo.vendor)
		if e != nil {
			return e
		}

		return c.EnsureLogIndex(ctx, task.ClusterName, task.EnvName, app)

	default:
	}

	return errors.Wrapf(_errcode.UnknownVendorError,
		"cluster(%s), env(%s), vendor(%s)", clusterInfo.name, clusterInfo.envName, clusterInfo.vendor)
}

// DeleteLogConfig 删除日志采集器配置
func (s *Service) DeleteLogConfig(ctx context.Context,
	clusterName entity.ClusterName, envName entity.AppEnvName, app *resp.AppDetailResp) error {
	clusterInfo, err := s.getClusterInfo(clusterName, string(envName))
	if err != nil {
		return err
	}

	// 删除时忽略日志采集器禁用状态, 因为删除时的禁用状态与创建时不一定一致

	switch clusterInfo.vendor {
	case entity.VendorAli:
		return s.DeleteAliLogConfig(ctx, clusterName, string(envName), &req.DeleteAliLogConfigReq{
			Namespace: string(envName),
			Name:      app.AliLogConfigName,
		})

	case entity.VendorHuawei:
		// 华为不删除AOM容器日志接入规则, 清理app时一并处理
		return nil

	default:
	}

	return errors.Wrapf(_errcode.UnknownVendorError,
		"cluster(%s), env(%s), vendor(%s)", clusterInfo.name, clusterInfo.envName, clusterInfo.vendor)
}

// DeleteLogStore 删除日志存储
func (s *Service) DeleteLogStore(ctx context.Context, task *resp.TaskDetailResp) error {
	clusterInfo, err := s.getClusterInfo(task.ClusterName, string(task.EnvName))
	if err != nil {
		return err
	}

	switch clusterInfo.vendor {
	case entity.VendorAli:
		// 兼容性逻辑
		// 最早阿里云 logstore 是每个应用关联一个, 但数量很快达到了上限并且无法提高到所需的量
		// 所以后来 logstore 改为了每个项目关联一个(对应 log_store_name 记录在 project 中)
		// 新部署的应用不再有自己的 logstore, 但需要兼容删除可能已经存在的老应用的 logstore
		if task.Param.CleanedAliLogStoreName == "" {
			return nil
		}

		getReq := &req.AliGetLogStoreDetailReq{
			ProjectName: config.Conf.Other.AliLogProjectStgName,
			StoreName:   task.Param.CleanedAliLogStoreName,
		}

		if task.EnvName == entity.AppEnvPrd || task.EnvName == entity.AppEnvPre {
			getReq.ProjectName = config.Conf.Other.AliLogProjectPrdName
		}

		_, err := s.GetLogStoreDetail(ctx, getReq)
		if err != nil {
			if errcode.EqualError(_errcode.AliResourceNotFoundError, err) {
				return nil
			}

			return err
		}

		return s.DeleteAliLogStore(ctx, &req.AliDeleteLogStoreReq{
			ProjectName: getReq.ProjectName,
			StoreName:   getReq.StoreName,
		})

	case entity.VendorHuawei:
		c, e := s.getVendorController(clusterInfo.vendor)
		if e != nil {
			return e
		}

		return c.DeleteLogConfig(ctx, task.ClusterName, task.EnvName, task.AppID)

	default:
	}

	return errors.Wrapf(_errcode.UnknownVendorError,
		"cluster(%s), env(%s), vendor(%s)", clusterInfo.name, clusterInfo.envName, clusterInfo.vendor)
}

// ApplyLogDump 同步日志转储配置
func (s *Service) ApplyLogDump(ctx context.Context,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp, task *resp.TaskDetailResp) error {
	clusterInfo, err := s.getClusterInfo(task.ClusterName, string(task.EnvName))
	if err != nil {
		return err
	}

	switch clusterInfo.vendor {
	case entity.VendorAli:
		return s.SyncAliLogStoreColdStorageShipper(ctx, task.ClusterName, &req.SyncAliLogStoreColdStorageShipperReq{
			EnvName:      task.EnvName,
			ProjectName:  project.Name,
			LogStoreName: project.LogStoreName,
		})

	case entity.VendorHuawei:
		// TODO: 华为云日志转储暂时没有提供删除功能, 暂不具备开放条件
		log.Warnc(ctx, "skipped log dump create phase for vendor(%s)", clusterInfo.vendor)
		return nil

	default:
	}

	return errors.Wrapf(_errcode.UnknownVendorError,
		"cluster(%s), env(%s), vendor(%s)", clusterInfo.name, clusterInfo.envName, clusterInfo.vendor)
}

// DeleteLogDump 删除日志转储配置
func (s *Service) DeleteLogDump(ctx context.Context,
	_ *resp.ProjectDetailResp, _ *resp.AppDetailResp, task *resp.TaskDetailResp) error {
	clusterInfo, err := s.getClusterInfo(task.ClusterName, string(task.EnvName))
	if err != nil {
		return err
	}

	switch clusterInfo.vendor {
	case entity.VendorAli:
		// 阿里云不需要删除日志转储
		return nil

	case entity.VendorHuawei:
		// TODO: 华为云日志转储暂时没有提供删除功能
		log.Warnc(ctx, "skipped log dump delete phase for vendor(%s)", clusterInfo.vendor)
		return nil

	default:
	}

	return errors.Wrapf(_errcode.UnknownVendorError,
		"cluster(%s), env(%s), vendor(%s)", clusterInfo.name, clusterInfo.envName, clusterInfo.vendor)
}
