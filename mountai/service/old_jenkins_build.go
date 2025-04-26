package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"rulai/dao"
	"rulai/models/entity"
	"rulai/utils"

	"gitlab.shanhai.int/sre/library/net/errcode"

	"gitlab.shanhai.int/sre/library/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Service) getProjectBuildAllValueFromCache(ctx context.Context, projectName string) ([]string, error) {
	return s.dao.GetProjectBuildAllValueFromCache(ctx, projectName)
}

func (s *Service) SyncOldJenkinsBuild(ctx context.Context) error {
	size := 30
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
			log.Errorc(ctx, "sync old jenkins error: %+v", err)
			return err
		}

		for _, project := range projects {
			vals, err := s.getProjectBuildAllValueFromCache(ctx, project.Name)
			if err != nil {
				log.Error("%s", err)
				continue
			}

			for _, val := range vals {
				now := time.Now()

				item := new(entity.JenkinsBuild)
				err = json.Unmarshal([]byte(val), item)
				if err != nil {
					log.Error("unmarshal %s %s", val, err)
					continue
				}

				createTime, err := time.Parse(utils.ImageTimeFormatLayout, item.CreateTime)
				if err != nil {
					log.Error("createTime %s %s", item.CreateTime, err)
					continue
				}

				build := &entity.JenkinsBuildImage{
					ID:           primitive.NewObjectID(),
					ProjectID:    project.ID,
					ProjectName:  project.Name,
					BuildID:      item.BuildID,
					JobName:      item.JobName,
					JobURL:       item.JobURL,
					ImageRepoURL: item.ImageRepoURL,
					BuildArg:     item.BuildArg,
					BranchName:   item.BranchName,
					CommitID:     item.CommitID,
					Description:  item.Description,
					UserID:       item.UserID,
					ImageTag:     fmt.Sprintf("%s-%s", item.CommitID, item.BranchName),
					CreateTime:   &createTime,
					UpdateTime:   &now,
				}

				_, err = s.dao.GetLastJenkinsBuildImage(ctx, bson.M{
					"branch_name": build.BranchName,
					"project_id":  build.ProjectID,
					"commit_id":   build.CommitID,
				})

				if err != nil {
					if errcode.EqualError(errcode.NoRowsFoundError, err) {
						err = s.dao.CreateJenkinsBuildImage(ctx, build)
						if err != nil {
							log.Error("insert image err %s", err)
						}
					} else {
						log.Error("find image err %s", err)
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
