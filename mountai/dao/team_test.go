package dao

import (
	"context"
	"testing"

	"rulai/models/entity"

	"github.com/stretchr/testify/assert"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestDao_Team(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		id, err := d.CreateSingleTeam(context.Background(), &entity.Team{
			ID:           primitive.NewObjectID(),
			Name:         "测试",
			DingHook:     "https://oapi.dingtalk.com/robot/send?access_token=93b0dd581161acafe876b833265591a3df96c87a8dba94d706bfbf94f1f49a1b",
			Label:        "unittest",
			AliAlarmName: "单元测试小组",
			ExtraDingHooks: map[string]string{
				"ci": "https://oapi.dingtalk.com/robot/send?access_token=93b0dd581161acafe876b833265591a3df96c87a8dba94d706bfbf94f1f49a1b",
			},
		})
		assert.Nil(t, err)

		team, err := d.FindSingleTeam(context.Background(), bson.M{
			"name": "测试",
		})
		assert.Nil(t, err)
		assert.Equal(t, "测试", team.Name)
		assert.Equal(t, "unittest", team.Label)

		err = d.UpdateSingleTeam(context.Background(), id.Hex(), bson.M{
			"$set": bson.M{
				"name": "测试小组",
			},
		})
		assert.Nil(t, err)

		count, err := d.CountTeam(context.Background(), bson.M{
			"label": "unittest",
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, count)

		teams, err := d.FindTeams(context.Background(), bson.M{
			"label": "unittest",
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(teams))
		assert.Equal(t, "测试小组", teams[0].Name)

		err = d.DeleteSingleTeam(context.Background(), bson.M{
			"_id": id,
		})
		assert.Nil(t, err)

		_, err = d.FindSingleTeam(context.Background(), bson.M{
			"_id": id,
		})
		assert.True(t, errcode.EqualError(errcode.NoRowsFoundError, err))
	})
}
