package service

import (
	"rulai/dao"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"

	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/reflect"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Service) GetResourceListFromCache(ctx context.Context, providerType entity.ProviderType,
	resourceType entity.ResourceType) (instances []*entity.ResourceInstance, err error) {
	return s.dao.GetResourceListFromCache(ctx, providerType, resourceType)
}

func (s *Service) GetProjectResources(ctx context.Context, projectID string, env entity.AppEnvName) (map[string]interface{}, error) {
	resource, err := s.dao.FindProjectResource(ctx, bson.M{
		"project_id": projectID,
		"env":        env,
	})
	if errcode.EqualError(errcode.NoRowsFoundError, err) {
		return make(map[string]interface{}), nil
	}
	if err != nil {
		return nil, err
	}

	ret, err := reflect.StructToMapByJson(resource)
	if err != nil {
		return nil, err
	}
	resourceMap, err := s.getResourceMapFromStruct(ctx, ret)

	if err != nil {
		return nil, err
	}

	for _, resourceType := range entity.AllResourceTypes {
		res := resourceMap[string(resourceType)]
		if len(res) > 0 {
			resources := make([]*entity.ResourceInstance, len(res))
			count := 0

			for idx := range res {
				resources[count] = res[idx]
				count++
			}
			ret[string(resourceType)] = resources
		}
	}

	return ret, nil
}

func (s *Service) UpdateProjectResources(ctx context.Context, projectID string, updateReq *req.UpdateProjectResourcesReq) error {
	changeMap, err := s.generateUpdateResourceFromReq(ctx, updateReq, false)
	if err != nil {
		return err
	}

	err = s.dao.UpsertProjectResource(ctx, bson.M{
		"project_id": projectID,
		"env":        updateReq.EnvName,
	}, bson.M{
		"$set": changeMap,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) generateUpdateResourceFromReq(ctx context.Context,
	updateReq *req.UpdateProjectResourcesReq, sync bool) (map[string]interface{}, error) {
	updateResourceMap, err := reflect.StructToMapByJson(updateReq)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}
	return s.generateUpdateResource(ctx, updateResourceMap, sync)
}

func (s *Service) generateUpdateResource(ctx context.Context,
	updateResourceMap map[string]interface{}, sync bool) (map[string]interface{}, error) {
	resourceMap, err := s.getResourceMapFromStruct(ctx, updateResourceMap)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}
	change := make(map[string]interface{})

	for _, resourceType := range entity.AllResourceTypes {
		if updateResourceMap[string(resourceType)] == nil {
			continue
		}
		ret := updateResourceMap[string(resourceType)].([]interface{})
		res := resourceMap[string(resourceType)]
		instanceIDs := make([]interface{}, 0)

		for _, id := range ret {
			instanceID := id.(string)
			if res[instanceID] == nil {
				if !sync {
					return nil, errors.Wrapf(errcode.InvalidParams, "resource: %s id: %s is not found.", resourceType, instanceID)
				}
			} else {
				instanceIDs = append(instanceIDs, id)
			}
		}
		if sync {
			change[string(resourceType)] = map[string]interface{}{
				"$each": instanceIDs,
			}
		} else {
			change[string(resourceType)] = instanceIDs
		}
	}

	return change, nil
}

func (s *Service) getResourceMapFromStruct(ctx context.Context,
	ret map[string]interface{}) (map[string]map[string]*entity.ResourceInstance, error) {
	resourceIDMap := make(map[string]map[string][]string)

	for _, resourceType := range entity.AllResourceTypes {
		if ret[string(resourceType)] == nil {
			continue
		}

		resourceIDs := ret[string(resourceType)].([]interface{})
		if len(resourceIDs) == 0 {
			continue
		}

		resourceIDMap[string(resourceType)] = make(map[string][]string)

		for _, resourceID := range resourceIDs {
			instanceID := resourceID.(string)

			for _, provider := range entity.AllProviderType {
				if !strings.HasPrefix(instanceID, string(provider)) {
					continue
				}

				if resourceIDMap[string(resourceType)][string(provider)] == nil {
					resourceIDMap[string(resourceType)][string(provider)] = make([]string, 0)
				}

				resourceIDMap[string(resourceType)][string(provider)] = append(resourceIDMap[string(resourceType)][string(provider)],
					strings.TrimPrefix(instanceID, fmt.Sprintf("%s_", provider)))

				break
			}
		}
	}

	resourceMap, err := s.dao.GetResourceMapFromCache(ctx, resourceIDMap)
	if err != nil {
		return nil, err
	}

	return resourceMap, nil
}

func (s *Service) SetLastSyncCommitID(ctx context.Context, commitID string) error {
	return s.dao.SetLastSyncCommitID(ctx, commitID)
}

func (s *Service) GetLastSyncCommitID(ctx context.Context) (string, error) {
	return s.dao.GetLastSyncCommitID(ctx)
}

func (s *Service) SyncResourcesFromConfigCenter(ctx context.Context, commitID string) error {
	size := 50
	afterID := ""

	for {
		limit := int64(size)
		projects, err := s.dao.FindProjects(ctx, bson.M{
			"_id": bson.M{
				"$gt": afterID,
			},
		}, &options.FindOptions{
			Limit: &limit,
			Sort:  dao.MongoSortByIDAsc,
		})
		if err != nil {
			log.Errorc(ctx, "sync config center resource error: %+v", err)
			return err
		}

		for _, project := range projects {
			for _, env := range entity.AppNormalEnvNames {
				ret, err := s.GetProjectResourceFromConfig(ctx, &req.GetProjectResourceFromConfigReq{
					CommitID: commitID,
					EnvName:  env,
				}, &resp.ProjectDetailResp{
					Name: project.Name,
				})
				if err != nil {
					log.Errorc(ctx, "sync config center resource error for project %s env %s: %+v", project.ID, env, err)
					continue
				}

				if len(ret.Mysql) == 0 && len(ret.Mongo) == 0 && len(ret.Redis) == 0 {
					continue
				}

				updateReq := &req.UpdateProjectResourcesReq{}
				if len(ret.Mysql) > 0 {
					updateReq.Rds = make([]string, len(ret.Mysql))
					for idx := range ret.Mysql {
						updateReq.Rds[idx] = fmt.Sprintf("%s_%s", entity.ProviderTypeAliyun, ret.Mysql[idx].ID)
					}
				}

				if len(ret.Mongo) > 0 {
					updateReq.Mongo = make([]string, len(ret.Mongo))
					for idx := range ret.Mongo {
						updateReq.Mongo[idx] = fmt.Sprintf("%s_%s", entity.ProviderTypeAliyun, ret.Mongo[idx].ID)
					}
				}

				if len(ret.Redis) > 0 {
					updateReq.Redis = make([]string, len(ret.Redis))
					for idx := range ret.Redis {
						updateReq.Redis[idx] = fmt.Sprintf("%s_%s", entity.ProviderTypeAliyun, ret.Redis[idx].ID)
					}
				}

				changeMap, err := s.generateUpdateResourceFromReq(ctx, updateReq, true)
				if err != nil {
					log.Errorc(ctx, "sync config center resource error for project %s env %s: %+v", project.ID, env, err)
					continue
				}

				err = s.dao.UpsertProjectResource(ctx, bson.M{
					"project_id": project.ID,
					"env":        env,
				}, bson.M{
					"$set": map[string]interface{}{
						"commit_id": commitID,
					},
					"$addToSet": changeMap,
				})

				if err != nil {
					log.Errorc(ctx, "sync config center resource error for project %s env %s: %+v", project.ID, env, err)
					continue
				}
			}
		}

		if len(projects) < size {
			break
		}

		afterID = projects[len(projects)-1].ID
	}

	return nil
}

func (s *Service) SyncResourcesFromApollo(ctx context.Context) error {
	size := 50
	afterID := ""

	for {
		limit := int64(size)
		projects, err := s.dao.FindProjects(ctx, bson.M{
			"apollo_appid": bson.M{
				"$gt": "",
			},
			"_id": bson.M{
				"$gt": afterID,
			},
		}, &options.FindOptions{
			Limit: &limit,
			Sort:  dao.MongoSortByIDAsc,
		})
		if err != nil {
			log.Errorc(ctx, "sync apollo resource error: %+v", err)
			return err
		}

		for _, project := range projects {
			for _, env := range []entity.AppEnvName{entity.AppEnvStg, entity.AppEnvPrd} {
				namespaces, err := s.dao.GetApolloNamespaceByAppID(ctx, project.ApolloAppID, env)
				if err != nil {
					log.Errorc(ctx, "sync apollo resource error for project: %s: %+v", project.ID, err)
					continue
				}

				for idx := range namespaces {
					err = s.syncProjectApolloConf(ctx, project, namespaces[idx], env)
					if err != nil {
						log.Errorc(ctx, "sync apollo resource error for project: %s namespace: %d: %+v", project.ID, namespaces[idx].ID, err)
					}
				}
			}
		}

		if len(projects) < size {
			break
		}

		afterID = projects[len(projects)-1].ID
	}

	return nil
}

func (s *Service) parseResourceConfig(ctx context.Context, confContent map[string]string) (map[string][]string, error) {
	resourceConf := make(map[string][]string)
	allResourceMap, err := s.preloadAllResourceInfo(ctx)
	if err != nil {
		return nil, err
	}

	for key := range confContent {
		for provider := range entity.ResourceConfParseMap {
			found := false

			for resourceType := range entity.ResourceConfParseMap[provider] {
				re := regexp.MustCompile(fmt.Sprintf(`([a-zA-Z0-9-]+)%s`, entity.ResourceConfParseMap[provider][resourceType]))
				res := re.FindStringSubmatch(confContent[key])
				if len(res) == 0 {
					continue
				}

				if resourceConf[string(resourceType)] == nil {
					resourceConf[string(resourceType)] = make([]string, 0)
				}

				// 匹配连接地址 (连接地址可以自定义)
				if resourceType == entity.ResourceTypeMongo || resourceType == entity.ResourceTypeRds ||
					resourceType == entity.ResourceTypeRedis {
					for _, instance := range allResourceMap[provider][resourceType] {
						if strings.Contains(instance.ConnectionStr, res[0]) {
							resourceConf[string(resourceType)] = append(resourceConf[string(resourceType)], instance.InstanceID)
							break
						}
					}
				} else {
					resourceConf[string(resourceType)] = append(resourceConf[string(resourceType)], fmt.Sprintf("%s_%s", provider, res[1]))
				}

				found = true

				break
			}

			if found {
				break
			}
		}
	}

	return resourceConf, nil
}

func (s *Service) preloadAllResourceInfo(ctx context.Context) (
	allResourceMap map[entity.ProviderType]map[entity.ResourceType][]*entity.ResourceInstance, err error) {
	allResourceMap = make(map[entity.ProviderType]map[entity.ResourceType][]*entity.ResourceInstance)

	for _, provider := range entity.AllProviderType {
		allResourceMap[provider] = make(map[entity.ResourceType][]*entity.ResourceInstance)

		for _, resourceType := range entity.AllResourceTypes {
			res, err := s.GetResourceListFromCache(ctx, provider, resourceType)
			if err != nil {
				return nil, err
			}

			allResourceMap[provider][resourceType] = res
		}
	}

	return allResourceMap, nil
}

func (s *Service) syncProjectApolloConf(ctx context.Context, project *entity.Project,
	namespace *entity.ApolloNamespace, env entity.AppEnvName) (err error) {
	confContent, err := s.getApolloConfig(ctx, namespace.AppID, namespace.ClusterName, namespace.NamespaceName, env)
	if err != nil {
		return err
	}

	resourceConf, err := s.parseResourceConfig(ctx, confContent)
	if err != nil {
		return err
	}

	updateResourceMap, err := reflect.StructToMapByJson(resourceConf)
	if err != nil {
		return err
	}

	changeMap, err := s.generateUpdateResource(ctx, updateResourceMap, true)
	if err != nil {
		return err
	}

	if len(changeMap) == 0 {
		return nil
	}

	err = s.dao.UpsertProjectResource(ctx, bson.M{
		"project_id": project.ID,
		"env":        env,
	}, bson.M{
		"$addToSet": changeMap,
	})
	if err != nil {
		return err
	}

	return nil
}
