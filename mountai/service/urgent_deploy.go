package service

import (
	"rulai/dao"
	"rulai/models/entity"

	"context"
	"fmt"
	"strings"
	"time"

	"gitlab.shanhai.int/sre/library/log"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	// Urgent deploy record template name.
	defaultUrgentDeployReportTmpName = "urgentDeploy"
	// Default send record time duration: 7 days.
	defaultRecordTime = 60 * 60 * 24 * 7
	// Default urgent deploy token.
	defaultUrgentDeployDingToken = "urgent_deploy"
	// Record title time layout.
	timeFormatLayout = "2006-01-02"
)

type urgentDeployRecord struct {
	Team        string
	ProjectName string
	Owners      string
	TotalTime   int
}

func (s *Service) ReportUrgentDeployRecord(ctx context.Context) error {
	records, err := s.dao.ListUrgentDeployDingTalkMsgRecords(
		ctx,
		bson.M{
			"env":         entity.AppEnvPrd,
			"is_p0_level": true,
			"create_time": bson.M{
				"$gte": time.Now().Add(-defaultRecordTime * time.Second),
			},
		},
		dao.MongoFindOptionWithSortByIDAsc,
	)
	if err != nil {
		return err
	}

	udRecords := make(map[string]*urgentDeployRecord)
	for _, record := range records {
		if r, ok := udRecords[record.ProjectID]; ok {
			r.TotalTime++
			continue
		}

		project, e := s.GetProjectDetail(ctx, record.ProjectID)
		if e != nil {
			log.Errorc(ctx, "project does not exist, project id: %s, err: %v", record.ProjectID, e)
			continue
		}

		ownersBuilder := strings.Builder{}
		for _, owner := range project.Owners {
			ownersBuilder.WriteString(fmt.Sprintf("%s ", owner.Name))
		}
		udRecords[record.ProjectID] = &urgentDeployRecord{
			Team:        project.Team.Label,
			ProjectName: project.Name,
			Owners:      ownersBuilder.String(),
			TotalTime:   1,
		}
	}

	robotMsg := "### {{.Title}}"
	for projectID, record := range udRecords {
		robotMsg = fmt.Sprintf("%s\n%s", robotMsg,
			fmt.Sprintf(">- 项目:%s, 次数:%d, 负责人:%s, 小组:%s, [项目地址](%s)", record.ProjectName,
				record.TotalTime, record.Owners, record.Team, s.GetAmsFrontendProjectURL(projectID, entity.AppEnvPrd)))
	}

	// title := fmt.Sprintf("P0服务紧急部署次数统计(%s~%s)", time.Now().Add(-defaultRecordTime*time.Second).Format(timeFormatLayout),
	// 	time.Now().Format(timeFormatLayout))
	// msgReq, err := s.generateRobotMessageReq(&req.AppOpMessage{
	// 	Title: title,
	// }, defaultUrgentDeployDingToken, defaultUrgentDeployReportTmpName, robotMsg)
	// if err != nil {
	// 	return err
	// }

	// if _, err := s.SendRobotMsgToDingTalk(ctx, msgReq); err != nil {
	// 	return errors.Wrap(err, "send robot message failed")
	// }

	return nil
}
