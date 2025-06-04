package abac

import (
	"context"
	"encoding/json"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository"
	"gitee.com/flycash/permission-platform/internal/repository/cache"
	"github.com/ego-component/eetcd"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type AttributeValueTask struct {
	repo        repository.AttributeValueRepository
	localCache  cache.ABACAttributeValCache
	redisCache  cache.ABACAttributeValCache
	resourceKey string
	subjectKey  string
	etcdClient  *eetcd.Component
}

func NewAttributeValueTask(repo repository.AttributeValueRepository, localCache, redisCache cache.ABACAttributeValCache, resourceKey, subjectKey string, etcdClient *eetcd.Component) *AttributeValueTask {
	return &AttributeValueTask{
		repo:        repo,
		localCache:  localCache,
		redisCache:  redisCache,
		resourceKey: resourceKey,
		subjectKey:  subjectKey,
		etcdClient:  etcdClient,
	}
}

func (t *AttributeValueTask) StartResLoop(ctx context.Context) error {
	return t.startLoop(ctx, t.resourceKey, t.updateResource)
}

func (t *AttributeValueTask) StartSubjectLoop(ctx context.Context) error {
	return t.startLoop(ctx, t.subjectKey, t.updateSubject)
}

func (t *AttributeValueTask) startLoop(ctx context.Context, key string,
	updateDataFunc func(ctx context.Context, val []byte) error,
) error {
	watchChan := t.etcdClient.Watch(ctx, key)
	for watchResp := range watchChan {
		for _, event := range watchResp.Events {
			if event.Type == clientv3.EventTypePut {
				// 更新热点数据
				_ = updateDataFunc(ctx, event.Kv.Value)
			}
		}
	}
	return nil
}

func (t *AttributeValueTask) updateSubject(ctx context.Context, value []byte) error {
	var obj domain.ABACObject
	err := json.Unmarshal(value, &obj)
	if err != nil {
		return err
	}
	obj, err = t.repo.FindSubjectValue(ctx, obj.BizID, obj.ID)
	if err != nil {
		return err
	}
	err = t.redisCache.SetAttrSubObj(ctx, []domain.ABACObject{obj})
	if err != nil {
		return err
	}
	return t.localCache.SetAttrSubObj(ctx, []domain.ABACObject{obj})
}

func (t *AttributeValueTask) updateResource(ctx context.Context, value []byte) error {
	var obj domain.ABACObject
	err := json.Unmarshal(value, &obj)
	if err != nil {
		return err
	}
	obj, err = t.repo.FindResourceValue(ctx, obj.BizID, obj.ID)
	if err != nil {
		return err
	}
	err = t.redisCache.SetAttrResObj(ctx, []domain.ABACObject{obj})
	if err != nil {
		return err
	}
	return t.localCache.SetAttrResObj(ctx, []domain.ABACObject{obj})
}
