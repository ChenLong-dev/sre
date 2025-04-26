package service

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"rulai/models/req"
)

func TestService_GetGitlabSingleProject(t *testing.T) {
	t.Run("424", func(t *testing.T) {
		res, err := s.GetGitlabSingleProject(context.Background(), "424")
		assert.Nil(t, err)
		assert.Equal(t, 424, res.ID)
	})
}

func TestService_GetGitlabProjectRawFile(t *testing.T) {
	t.Run("1647", func(t *testing.T) {
		res, err := s.GetGitlabProjectRawFile(context.Background(), "1647", "master",
			".gitlab/issue_templates/bug.md")
		assert.Nil(t, err)
		fmt.Println(res)
	})
}

func TestService_GetQTFrameworkVersion(t *testing.T) {
	t.Run("1647", func(t *testing.T) {
		res, err := s.GetQTFrameworkVersion(context.Background(), "1647", "master")
		assert.Nil(t, err)
		fmt.Printf("%#v\n", res)
	})

	t.Run("1449", func(t *testing.T) {
		res, err := s.GetQTFrameworkVersion(context.Background(), "1449", "master")
		assert.Nil(t, err)
		fmt.Printf("%#v\n", res)
	})

	t.Run("1489", func(t *testing.T) {
		res, err := s.GetQTFrameworkVersion(context.Background(), "1489", "master")
		assert.Nil(t, err)
		fmt.Printf("%#v\n", res)
	})
}

func Test_Service_GetGitlabUser(t *testing.T) {
	ctx := context.Background()
	t.Run("git user 236", func(t *testing.T) {
		userID := 236
		res, err := s.GetGitlabUser(ctx, strconv.Itoa(userID))
		assert.Nil(t, err)
		assert.Equal(t, userID, res.ID)
		assert.Equal(t, res.Name, "王晋元")
		assert.Equal(t, res.Username, "wangjinyuan")
	})
}

func TestService_GetGitlabProjectMembers(t *testing.T) {
	t.Run("389", func(t *testing.T) {
		projectID := "389"
		onePageResult, err := s.GetGitlabProjectMembers(context.Background(), projectID, &req.GetProjectMembersReq{
			Page:     1,
			PageSize: 30,
		})
		assert.Nil(t, err)
		if len(onePageResult) >= 20 {
			res, err := s.GetGitlabProjectAllActiveMembers(context.Background(), projectID)
			assert.Nil(t, err)
			assert.True(t, len(res) >= 20)
			fmt.Printf("%#v\n", res)
		}
	})
}
