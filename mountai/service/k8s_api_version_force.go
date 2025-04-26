package service

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"rulai/models/entity"
	"rulai/utils"
)

// 强制指定的 k8s 资源版本
var (
	groupVersionExtensionsV1beta1 = &schema.GroupVersion{Group: entity.GroupNameExtensions, Version: "v1beta1"}
	groupVersionAppsV1            = &schema.GroupVersion{Group: "apps", Version: "v1"}
)

// k8sObjectKindsForceToExtensionsV1beta1OrAppsV1 根据 server 版本, 强制使用 extensions/v1beta1 或 apps/v1 版本的 k8s 资源类型
var k8sObjectKindsForceToExtensionsV1beta1OrAppsV1 = []string{
	entity.K8sObjectKindDeployment,
	entity.K8sObjectKindReplicaSet,
}

// k8sObjectKindsForceToExtensionsV1beta1 无视 server 版本, 强制使用 extensions/v1beta1 版本的 k8s 资源类型
var k8sObjectKindsForceToExtensionsV1beta1 = []string{
	// TODO: why need v1beta1?
	// entity.K8sObjectKindIngress,
}

// forceK8sObjectKinds 强制指定 k8s 资源版本为先前使用的版本
func forceK8sObjectKinds(cluster *k8sCluster) error {
	minorClusterVersionStr := utils.UnifyK8sMinorVersion(cluster.version.Minor)
	minorClusterVersion, err := strconv.Atoi(minorClusterVersionStr)
	if err != nil {
		return errors.Wrapf(errcode.InternalError, "unknown minor version(%s)", minorClusterVersionStr)
	}

	if cluster.version.Major == "1" {
		for _, kind := range k8sObjectKindsForceToExtensionsV1beta1OrAppsV1 {
			gv := cluster.k8sGroupVersions[kind]
			if minorClusterVersion < 18 {
				if !equalK8sGroupVersion(gv, groupVersionExtensionsV1beta1) {
					log.Infoc(context.Background(),
						"Forced %s's apiVersion(cluster=%s, env=%s) from %s to: %s",
						kind, cluster.name, cluster.envName, gv.String(), groupVersionExtensionsV1beta1.String())
					cluster.k8sGroupVersions[kind] = groupVersionExtensionsV1beta1
				}
			} else if !equalK8sGroupVersion(gv, groupVersionAppsV1) {
				log.Infoc(context.Background(),
					"Forced %s's apiVersion(cluster=%s, env=%s) from %s to: %s",
					kind, cluster.name, cluster.envName, gv.String(), groupVersionAppsV1.String())
				cluster.k8sGroupVersions[kind] = groupVersionAppsV1
			}
		}
	}

	for _, kind := range k8sObjectKindsForceToExtensionsV1beta1 {
		gv := cluster.k8sGroupVersions[kind]
		if !equalK8sGroupVersion(gv, groupVersionExtensionsV1beta1) {
			log.Infoc(context.Background(),
				"Forced %s' apiVersion(cluster=%s, env=%s) from %s to: %s",
				kind, cluster.name, cluster.envName, gv.String(), groupVersionExtensionsV1beta1.String())
			cluster.k8sGroupVersions[kind] = groupVersionExtensionsV1beta1
		}
	}

	return nil
}
