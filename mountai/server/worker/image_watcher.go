package worker

import (
	"context"
	"time"

	framework "gitlab.shanhai.int/sre/app-framework"
	"gitlab.shanhai.int/sre/library/goroutine"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/sentry"

	"rulai/models/entity"
	"rulai/models/req"
	"rulai/service"
)

const defaultPodWatchTimeout = time.Second * 5

// GetPodWatcher pod 监听器
// 当前主要是用于监控是否有 pod 使用了非华为云的镜像, 避免华为云访问外网镜像导致各种失败
func GetPodWatcher() framework.ServerInterface {
	svr := new(framework.JobServer)
	svr.SetJob("pod watcher", func(ctx context.Context) error {
		tmpCtx, cancel := context.WithTimeout(ctx, defaultPodWatchTimeout)
		defer cancel()

		clusters, err := service.SVC.GetClusters(tmpCtx, new(req.GetClustersReq))
		if err != nil {
			return err
		}

		var ok bool
		wg := goroutine.New("pod-watcher")
		m := make(map[entity.ClusterName]struct{}, len(clusters))
		for _, cluster := range clusters {
			// clusters 是按照 clusterName 和 envName 聚合, 存在重复
			if _, ok = m[cluster.Name]; ok {
				continue
			}

			curCluster := cluster
			m[curCluster.Name] = struct{}{}

			wg.Go(ctx, string(curCluster.Name), func(ctx context.Context) error {
				e := service.SVC.WatchClusterPods(ctx, curCluster.Name, curCluster.Env)
				if e != nil {
					log.Errorc(ctx, "watch pods for cluster(%s) and env(%s) failed: %s", curCluster.Name, curCluster.Env, e)
					// 监听失败的时候需要有 sentry 报警以便感知
					sentry.CaptureWithBreadAndTags(ctx, e, &sentry.Breadcrumb{
						Category: "WatchClusterPods",
						Data: map[string]interface{}{
							"ClusterName": curCluster.Name,
							"Env":         curCluster.Env,
						},
					})

					return e
				}

				return nil
			})
		}

		err = wg.Wait()
		if err != nil {
			log.Errorc(ctx, "watch pods ended with error: %s", err)
		}

		return nil
	})

	return svr
}
