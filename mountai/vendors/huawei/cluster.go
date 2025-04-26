package huawei

import (
	"github.com/pkg/errors"

	"rulai/models/entity"
	_errcode "rulai/utils/errcode"
)

// getClusterConfig 获取运营商内的k8s集群配置
func (c *Controller) getClusterConfig(envName entity.AppEnvName, clusterName entity.ClusterName) (*ClusterConfig, error) {
	if len(c.Clusters) == 0 {
		return nil, errors.Wrapf(_errcode.ClusterNotExistsInVendorError, "vendor=%s, env=%s, cluster=%s", c.Name(), envName, clusterName)
	}

	clustersInEnv, ok := c.Clusters[envName]
	if !ok || clustersInEnv == nil {
		return nil, errors.Wrapf(_errcode.ClusterNotExistsInVendorError, "vendor=%s, env=%s, cluster=%s", c.Name(), envName, clusterName)
	}

	clusterInfo, ok := clustersInEnv[clusterName]
	if !ok || clusterInfo == nil {
		return nil, errors.Wrapf(_errcode.ClusterNotExistsInVendorError, "vendor=%s, env=%s, cluster=%s", c.Name(), envName, clusterName)
	}

	return clusterInfo, nil
}
