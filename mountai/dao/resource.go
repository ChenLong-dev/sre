package dao

import (
	"rulai/models/entity"

	"context"
	"encoding/json"
	"fmt"
	"time"

	u "github.com/alxrm/ugo"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const LastSyncCommitIDKey = "last_sync_commit"
const LastSyncCommitIDTTL = 60 * 60 * 24
const ResourceTTL = 60 * 60 * 24

func (d *Dao) getResourceCacheKey(providerType, resourceType string) string {
	return fmt.Sprintf("%s%s:%s%s", entity.ProviderTypePrefix, providerType, entity.ResourceTypeCachePrefix, resourceType)
}

func (d *Dao) GetResourceKeysFromCache(ctx context.Context, providerType entity.ProviderType,
	resourceType entity.ResourceType) (instanceIDs []string, err error) {
	con := d.Redis.Get()
	defer con.Close()

	value, err := redis.Values(
		con.Do(
			ctx, "hkeys",
			d.getResourceCacheKey(string(providerType), string(resourceType)),
		),
	)

	if err != nil {
		return nil, errors.Wrapf(errcode.RedisError, err.Error())
	}

	if value == nil {
		return instanceIDs, nil
	}

	if err = redis.ScanSlice(value, &instanceIDs); err != nil {
		return nil, errors.Wrapf(errcode.InternalError, err.Error())
	}

	return instanceIDs, nil
}

func (d *Dao) SetResourceListToCache(ctx context.Context, providerType entity.ProviderType,
	resourceType entity.ResourceType, instances []*entity.ResourceInstance) (err error) {
	key := d.getResourceCacheKey(string(providerType), string(resourceType))
	instanceIDs, err := d.GetResourceKeysFromCache(ctx, providerType, resourceType)

	if err != nil {
		return err
	}

	instanceSeq := u.From(instanceIDs, len(instanceIDs))
	hmsetString := redis.Args{}.Add(key)

	for _, instance := range instances {
		hmsetString = hmsetString.Add(instance.ID)

		var ret []byte

		ret, err = json.Marshal(instance)
		if err != nil {
			return errors.Wrap(errcode.InternalError, err.Error())
		}

		hmsetString = hmsetString.Add(string(ret))

		instanceSeq = u.Without(instanceSeq, instance.ID, nil)
	}

	con := d.Redis.Get()
	defer con.Close()

	err = con.Send(ctx, "hmset", hmsetString...)

	if err != nil {
		return errors.Wrapf(errcode.RedisError, err.Error())
	}

	if len(instanceSeq) > 0 {
		for _, instanceID := range instanceSeq {
			err = con.Send(ctx, "hdel", key, instanceID)
			if err != nil {
				return errors.Wrapf(errcode.RedisError, err.Error())
			}
		}
	}

	_, err = con.Do(ctx, "expire", key, ResourceTTL)
	if err != nil {
		return errors.Wrapf(errcode.RedisError, err.Error())
	}

	return nil
}

func (d *Dao) GetResourceListFromCache(ctx context.Context, providerType entity.ProviderType,
	resourceType entity.ResourceType) (instances []*entity.ResourceInstance, err error) {
	con := d.Redis.Get()
	defer con.Close()

	value, err := redis.Values(con.Do(ctx, "hgetall", d.getResourceCacheKey(string(providerType), string(resourceType))))

	if err != nil {
		return nil, errors.Wrapf(errcode.RedisError, err.Error())
	}

	if value == nil {
		return instances, nil
	}

	var res [][]byte
	if err = redis.ScanSlice(value, &res); err != nil {
		return nil, errors.Wrapf(errcode.InternalError, err.Error())
	}

	for idx := range res {
		if idx%2 == 0 {
			continue
		}

		instance := &entity.ResourceInstance{}

		err = json.Unmarshal(res[idx], instance)
		if err != nil {
			return nil, errors.Wrap(errcode.InternalError, err.Error())
		}

		instances = append(instances, instance)
	}

	return instances, nil
}

func (d *Dao) GetResourceMapFromCache(ctx context.Context,
	resourceIDMap map[string]map[string][]string) (resourceMap map[string]map[string]*entity.ResourceInstance, err error) {
	con := d.Redis.Get()
	defer con.Close()

	count := 0

	for resourceType := range resourceIDMap {
		for providerType := range resourceIDMap[resourceType] {
			count++

			err = con.Send(
				ctx, "hmget",
				redis.Args{}.Add(d.getResourceCacheKey(providerType, resourceType)).
					AddFlat(resourceIDMap[resourceType][providerType])...,
			)

			if err != nil {
				return nil, errors.Wrapf(errcode.RedisError, err.Error())
			}
		}
	}

	err = con.Flush(ctx)
	if err != nil {
		return nil, errors.Wrapf(errcode.RedisError, err.Error())
	}

	resourceMap = make(map[string]map[string]*entity.ResourceInstance)

	for i := 0; i < count; i++ {
		value, err := redis.Values(con.Receive(ctx))

		if err != nil {
			return nil, errors.Wrapf(errcode.RedisError, err.Error())
		}

		if value == nil {
			continue
		}

		var res [][]byte
		if err = redis.ScanSlice(value, &res); err != nil {
			return nil, errors.Wrapf(errcode.InternalError, err.Error())
		}

		for idx := range res {
			if res[idx] == nil {
				continue
			}

			instance := &entity.ResourceInstance{}

			err = json.Unmarshal(res[idx], instance)
			if err != nil {
				return nil, errors.Wrap(errcode.InternalError, err.Error())
			}

			if resourceMap[string(instance.Type)] == nil {
				resourceMap[string(instance.Type)] = make(map[string]*entity.ResourceInstance)
			}

			resourceMap[string(instance.Type)][instance.InstanceID] = instance
		}
	}

	return resourceMap, nil
}

func (d *Dao) FindProjectResource(ctx context.Context, filter bson.M) (*entity.Resource, error) {
	resource := new(entity.Resource)

	err := d.Mongo.ReadOnlyCollection(resource.TableName()).
		FindOne(ctx, filter).
		Decode(resource)
	if err == mongo.ErrNoDocuments {
		return nil, errors.Wrapf(errcode.NoRowsFoundError, "%s", err)
	} else if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return resource, nil
}

func (d *Dao) UpsertProjectResource(ctx context.Context, filter, change bson.M) error {
	change["$currentDate"] = bson.M{
		"update_time": true,
	}

	change["$setOnInsert"] = bson.M{
		"create_time": time.Now(),
	}
	_, err := d.Mongo.Collection(new(entity.Resource).TableName()).
		UpdateOne(ctx, filter, change, options.Update().SetUpsert(true))

	if err != nil {
		return errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return nil
}

func (d *Dao) SetLastSyncCommitID(ctx context.Context, commitID string) error {
	con := d.Redis.Get()
	defer con.Close()
	_, err := con.Do(ctx, "setex", redis.Args{}.Add(LastSyncCommitIDKey).Add(LastSyncCommitIDTTL).Add(commitID)...)

	if err != nil {
		return errors.Wrapf(errcode.RedisError, err.Error())
	}

	return nil
}

func (d *Dao) GetLastSyncCommitID(ctx context.Context) (string, error) {
	con := d.Redis.Get()
	defer con.Close()

	value, err := redis.String(con.Do(ctx, "get", LastSyncCommitIDKey))

	if err == redis.ErrNil {
		return "", nil
	}

	if err != nil {
		return "", errors.Wrapf(errcode.RedisError, err.Error())
	}

	return value, nil
}
