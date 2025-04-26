package job

import (
	"rulai/config"
	"rulai/service"

	"context"

	framework "gitlab.shanhai.int/sre/app-framework"
	"gitlab.shanhai.int/sre/library/log"
)

func GetSyncConfigCenterResourceServer() framework.ServerInterface {
	svr := new(framework.JobServer)

	// ====================
	// >>>请勿删除<<<
	//
	// 根据实际情况修改
	// ====================
	// 设置任务函数
	svr.SetJob("sync_config_center_resource", func(ctx context.Context) error {
		branch, err := service.SVC.GetGitlabSingleProjectBranch(
			ctx,
			config.Conf.Other.ConfigCenter.ProjectID,
			config.Conf.Other.ConfigCenter.Branch,
		)
		if err != nil {
			return err
		}
		commitID := branch.Commit.ShortID

		lastSyncCommitID, err := service.SVC.GetLastSyncCommitID(ctx)
		if err != nil {
			return err
		}
		if lastSyncCommitID == commitID {
			log.Infoc(ctx, "last sync commit id: %s, no sync now", commitID)
			return nil
		}

		err = service.SVC.SyncResourcesFromConfigCenter(ctx, commitID)
		if err != nil {
			return err
		}

		err = service.SVC.SetLastSyncCommitID(ctx, commitID)
		if err != nil {
			return err
		}

		return nil
	})

	return svr
}
