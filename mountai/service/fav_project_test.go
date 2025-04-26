package service

import (
	"rulai/models/req"

	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_FavProject(t *testing.T) {
	res, err := s.CreateFavProject(context.Background(), &req.CreateFavProjectReq{
		ProjectID: "276",
	}, "171")
	assert.Nil(t, err)
	assert.Equal(t, "171", res.User.ID)
	assert.Equal(t, "276", res.Project.ID)

	list, err := s.GetFavProjects(context.Background(), &req.GetFavProjectReq{
		UserID:    "171",
		ProjectID: "276",
	})
	assert.Nil(t, err)
	assert.Equal(t, list[0].User.ID, "171")
	assert.Equal(t, list[0].Project.ID, "276")

	err = s.DeleteFavProject(context.Background(), res.User.ID, res.Project.ID)
	assert.Nil(t, err)

	list, err = s.GetFavProjects(context.Background(), &req.GetFavProjectReq{
		UserID:    "171",
		ProjectID: "276",
	})
	assert.Nil(t, err)
	assert.Len(t, list, 0)
}
