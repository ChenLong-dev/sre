package service

import (
	"rulai/config"
	"rulai/dao"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"

	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	p0DeployDingToken = "p0_deploy"
	defaultTmpName    = "appOpCropMsg"
	defaultAppOpMsg   = `
### {{.Title}}
>- [项目名] {{.ProjectName}}
>- [应用名] {{.AppName}}
>- [环境名] {{.Env}}
>- [操作人] {{.UserName}}
>- [操作] {{.Action}}
>- [时间] {{.OpTime}}
#### [进入AMS, 查看详情]({{.DetailURL}})
`
)

// MessageOperator represents a message queue event operator.
type MessageOperator interface {
	// Check if we should handle this event
	ShouldHandle(ctx context.Context) (bool, error)
	// Handle this event
	Handle(ctx context.Context) error
}

// P0AppOperation represents operations for P0 app
type P0AppOperation struct {
	*Service
	message *entity.SubscribeEventMsg
	project *resp.ProjectDetailResp
	app     *resp.AppDetailResp
	task    *resp.TaskDetailResp
	user    *resp.UserProfileResp
}

func newP0AppOperation(service *Service, message *entity.SubscribeEventMsg, project *resp.ProjectDetailResp,
	app *resp.AppDetailResp, task *resp.TaskDetailResp, user *resp.UserProfileResp) MessageOperator {
	return &P0AppOperation{
		Service: service,
		message: message,
		project: project,
		app:     app,
		task:    task,
		user:    user,
	}
}

// ShouldHandle implements MessageOperator
func (p *P0AppOperation) ShouldHandle(_ context.Context) (bool, error) {
	if p.message.Env != entity.AppEnvPrd || (p.task.Action != entity.TaskActionFullCanaryDeploy && p.task.Action != entity.TaskActionFullDeploy) ||
		p.task.Status != entity.TaskStatusSuccess {
		return false, nil
	}

	return p.IsP0LevelProject(p.project), nil
}

// Handle implements MessageOperator
func (p *P0AppOperation) Handle(ctx context.Context) error {
	msgReq, err := p.generateRobotMessageReq(&req.AppOpMessage{
		Title:           fmt.Sprintf("%s 进行了%s操作", p.project.Name, entity.GetSubscribeActionDisplay(p.message.ActionType)),
		ProjectName:     p.project.Name,
		AppName:         p.app.Name,
		UserName:        p.user.Name,
		Env:             string(p.message.Env),
		Action:          entity.GetSubscribeActionDisplay(p.message.ActionType),
		OpTime:          p.message.OpTime,
		DetailURL:       p.GetAmsFrontendProjectURL(p.project.ID, p.message.Env),
		ProjectLanguage: p.project.Language,
		TeamName:        p.project.Team.Name,
		Branch:          p.task.GetDeployBranch(),
	})
	if err != nil {
		return errors.Wrapf(err, "send p0 operation msg fail, task id: %s", p.message.TaskID)
	}

	if err = p.SendRobotMsgToFeishu(ctx, msgReq); err != nil {
		return errors.Wrapf(err, "send p0 operation msg fail, task id: %s", p.message.TaskID)
	}
	return nil
}

// UserSubAppOperation represents operations which user subscribes
type UserSubAppOperation struct {
	*Service
	message *entity.SubscribeEventMsg
	project *resp.ProjectDetailResp
	app     *resp.AppDetailResp
	task    *resp.TaskDetailResp
	user    *resp.UserProfileResp

	userIDs []string
	emails  []string
}

func newUserSubAppOperation(service *Service, message *entity.SubscribeEventMsg, project *resp.ProjectDetailResp,
	app *resp.AppDetailResp, task *resp.TaskDetailResp, user *resp.UserProfileResp) MessageOperator {
	return &UserSubAppOperation{
		Service: service,
		message: message,
		project: project,
		app:     app,
		task:    task,
		user:    user,
		userIDs: make([]string, 0),
		emails:  make([]string, 0),
	}
}

// ShouldHandle implements MessageOperator
func (u *UserSubAppOperation) ShouldHandle(ctx context.Context) (bool, error) {
	if u.task.Status != entity.TaskStatusSuccess {
		return false, nil
	}

	appID, err := primitive.ObjectIDFromHex(u.message.AppID)
	if err != nil {
		return false, errors.Wrapf(_errcode.InvalidHexStringError, "%s", err)
	}
	subscriptions, err := u.dao.FindUserSubscriptionList(
		ctx,
		bson.M{
			"app_id":      appID,
			"env":         u.message.Env,
			"action_type": u.message.ActionType,
		},
		dao.MongoFindOptionWithSortByIDAsc,
	)
	if err != nil {
		return false, err
	}

	if len(subscriptions) == 0 {
		return false, nil
	}

	for _, subscription := range subscriptions {
		u.userIDs = append(u.userIDs, subscription.UserID)
	}

	users, err := u.dao.FindUserAuth(
		ctx,
		bson.M{
			"_id": bson.M{"$in": u.userIDs},
		},
		dao.MongoFindOptionWithSortByIDAsc,
	)
	if err != nil {
		return false, err
	}

	for _, user := range users {
		u.emails = append(u.emails, user.Email)
	}

	return len(u.emails) != 0, nil
}

// Handle implements MessageOperator
func (u *UserSubAppOperation) Handle(ctx context.Context) error {
	res, err := u.SendDingCropMessage(ctx, &req.AppOpMessage{
		Title:       fmt.Sprintf("%s 进行了%s操作", u.project.Name, entity.GetTaskActionDisplay(u.task.Action)),
		ProjectName: u.project.Name,
		AppName:     u.app.Name,
		Env:         string(u.message.Env),
		Action:      entity.GetTaskActionDisplay(u.task.Action),
		OpTime:      u.message.OpTime,
		UserName:    u.user.Name,
		DetailURL:   u.GetAmsFrontendProjectURL(u.project.ID, u.message.Env),
	}, u.emails, defaultTmpName, defaultAppOpMsg)
	if err != nil {
		return err
	}

	appID, err := primitive.ObjectIDFromHex(u.message.AppID)
	if err != nil {
		return errors.Wrapf(_errcode.InvalidHexStringError, "%s", err)
	}

	now := time.Now()
	return u.CreateDingTalkMsgRecord(ctx, &entity.DingTalkMessageRecord{
		ID:      primitive.NewObjectID(),
		UserIDs: u.userIDs,
		Content: entity.DingTalkMessageContent{
			ActionType: u.message.ActionType,
			AppID:      appID,
			ProjectID:  u.app.ProjectID,
			Env:        u.message.Env,
		},
		TaskID:     res.TaskID,
		CreateTime: &now,
		UpdateTime: &now,
	})
}

// FailedTaskOperation represents operations for failed task
type FailedTaskOperation struct {
	*Service
	message *entity.SubscribeEventMsg
	project *resp.ProjectDetailResp
	app     *resp.AppDetailResp
	task    *resp.TaskDetailResp
	user    *resp.UserProfileResp
}

func newFailedTaskOperation(service *Service, message *entity.SubscribeEventMsg, project *resp.ProjectDetailResp,
	app *resp.AppDetailResp, task *resp.TaskDetailResp, user *resp.UserProfileResp) MessageOperator {
	return &FailedTaskOperation{
		Service: service,
		message: message,
		project: project,
		app:     app,
		task:    task,
		user:    user,
	}
}

// ShouldHandle implements MessageOperator
func (f *FailedTaskOperation) ShouldHandle(ctx context.Context) (bool, error) {
	return f.task.Status == entity.TaskStatusFail, nil
}

// Handle implements MessageOperator
func (f *FailedTaskOperation) Handle(ctx context.Context) error {
	res, err := f.SendDingCropMessage(ctx, &req.AppOpMessage{
		Title: fmt.Sprintf("%s的应用%s%s失败", f.project.Name,
			f.app.Name, entity.GetTaskActionDisplay(f.task.Action)),
		ProjectName: f.project.Name,
		AppName:     f.app.Name,
		Env:         string(f.message.Env),
		Action:      entity.GetTaskActionDisplay(f.task.Action),
		OpTime:      f.message.OpTime,
		UserName:    f.user.Name,
		DetailURL:   f.GetAmsFrontendProjectURL(f.project.ID, f.message.Env),
	}, []string{f.user.Email}, defaultTmpName, defaultAppOpMsg)
	if err != nil {
		return err
	}

	appID, err := primitive.ObjectIDFromHex(f.message.AppID)
	if err != nil {
		return errors.Wrapf(_errcode.InvalidHexStringError, "%s", err)
	}

	now := time.Now()
	return f.CreateDingTalkMsgRecord(ctx, &entity.DingTalkMessageRecord{
		ID:      primitive.NewObjectID(),
		UserIDs: []string{f.user.ID},
		Content: entity.DingTalkMessageContent{
			ActionType: f.message.ActionType,
			AppID:      appID,
			ProjectID:  f.app.ProjectID,
			Env:        f.message.Env,
		},
		TaskID:     res.TaskID,
		CreateTime: &now,
		UpdateTime: &now,
	})
}

// HandleAppOpMsgEvent handle message event
func (s *Service) HandleAppOpMsgEvent(ctx context.Context, msg *sarama.ConsumerMessage) error {
	message := new(entity.SubscribeEventMsg)
	err := json.Unmarshal(msg.Value, message)
	if err != nil {
		return errors.Wrap(errcode.InternalError, err.Error())
	}

	app, err := s.GetAppDetail(ctx, message.AppID)
	if err != nil {
		return err
	}

	project, err := s.GetProjectDetail(ctx, app.ProjectID)
	if err != nil {
		return err
	}

	user, err := s.GetUserInfo(ctx, message.OperatorID)
	if err != nil {
		return err
	}

	task, err := s.GetTaskDetail(ctx, message.TaskID)
	if err != nil {
		return err
	}

	operationMsgEvents := []MessageOperator{
		newP0AppOperation(s, message, project, app, task, user),
		// newUserSubAppOperation(s, message, project, app, task, user),
		// newFailedTaskOperation(s, message, project, app, task, user),
	}
	for i := range operationMsgEvents {
		shouldHandle, err := operationMsgEvents[i].ShouldHandle(ctx)
		if err != nil {
			log.Errorc(ctx, "%s", err)
			continue
		}
		if !shouldHandle {
			continue
		}
		if err = operationMsgEvents[i].Handle(ctx); err != nil {
			log.Errorc(ctx, "%s", err)
		}
	}

	return nil
}

// PublishAppOpEvent 推送订阅事件
func (s *Service) PublishAppOpEvent(ctx context.Context, msg *entity.SubscribeEventMsg) error {
	value, err := json.Marshal(msg)
	if err != nil {
		return errors.Wrap(errcode.InternalError, err.Error())
	}

	sendMsg := &sarama.ProducerMessage{
		Topic: config.Conf.AppOpConsumer.Topic,
		Key:   sarama.StringEncoder(msg.AppID),
		Value: sarama.StringEncoder(value),
	}

	partition, offset, err := s.dao.KafkaProducer.SendMessage(sendMsg)
	if err != nil {
		return errors.Wrap(errcode.KafkaError, err.Error())
	}

	log.Infoc(ctx, "publish app op message success: topic[%q], partition[%d], offset[%d], msg[%s]",
		config.Conf.AppOpConsumer.Topic, partition, offset, string(value))

	return nil
}

// UserSubscribe 订阅
func (s *Service) UserSubscribe(ctx context.Context, subscribeReq *req.UserSubscribeReq, userID string) (err error) {
	appID, err := primitive.ObjectIDFromHex(subscribeReq.AppID)
	if err != nil {
		return errors.Wrapf(_errcode.InvalidHexStringError, "%s", err)
	}

	now := time.Now()
	upsert := true

	return s.dao.UpdateUserSubscription(ctx, bson.M{
		"app_id":      appID,
		"user_id":     userID,
		"env":         subscribeReq.EnvName,
		"action_type": subscribeReq.Action,
	}, bson.M{
		"$set": bson.M{
			"app_id":      appID,
			"user_id":     userID,
			"env":         subscribeReq.EnvName,
			"action_type": subscribeReq.Action,
			"update_time": now,
		},
		"$setOnInsert": bson.M{"create_time": now},
	}, &options.UpdateOptions{
		Upsert: &upsert,
	})
}

// UserUnsubscribe 取消订阅
func (s *Service) UserUnsubscribe(ctx context.Context, unsubscribeReq *req.UserUnsubscribeReq, userID string) (err error) {
	appID, err := primitive.ObjectIDFromHex(unsubscribeReq.AppID)
	if err != nil {
		return errors.Wrapf(_errcode.InvalidHexStringError, "%s", err)
	}

	return s.dao.DeleteUserSubscription(ctx, bson.M{
		"app_id":      appID,
		"env":         unsubscribeReq.EnvName,
		"action_type": unsubscribeReq.Action,
		"user_id":     userID,
	})
}

// GetUserSubscribeInfo 获取订阅信息
func (s *Service) GetUserSubscribeInfo(ctx context.Context, env entity.AppEnvName,
	appID, userID string) (subscriptions []string, err error) {
	appObjectID, err := primitive.ObjectIDFromHex(appID)
	if err != nil {
		return nil, errors.Wrapf(_errcode.InvalidHexStringError, "%s", err)
	}

	subscriptions = make([]string, 0)

	subs, err := s.dao.FindUserSubscriptionList(
		ctx,
		bson.M{
			"app_id":  appObjectID,
			"user_id": userID,
			"env":     env,
		},
		dao.MongoFindOptionWithSortByIDAsc,
	)
	if err != nil {
		return nil, err
	}

	for _, sub := range subs {
		subscriptions = append(subscriptions, string(sub.ActionType))
	}

	return subscriptions, nil
}
