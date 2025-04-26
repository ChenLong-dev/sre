package huawei

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/sdkerr"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/swr/v2/model"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"

	"rulai/models/req"
	"rulai/models/resp"
)

const ImageTimeLayout = "2006-01-02T15:04:05Z"

func (c *Controller) GetRepoTags(ctx context.Context, getReq *req.GetRepoTagsReq) ([]*resp.GetDockerTagsResp, int, error) {
	limit := strconv.Itoa(getReq.Size)
	offset := "0"
	orderColumn := "updated_at"
	namespace := "qt-apps"
	orderType := model.GetListRepositoryTagsRequestOrderTypeEnum().DESC

	if getReq.Page-1 > 0 {
		offset = strconv.Itoa((getReq.Page - 1) * getReq.Size)
	}

	repo, err := c.SwrClient.ShowRepository(&model.ShowRepositoryRequest{
		Namespace:  namespace,
		Repository: getReq.ProjectName,
	})
	if err != nil {
		// 如果仓库不存在,返回空结果集,而不是报错信息,否则会导致 Jenkins 打包失败
		if responseError, ok := err.(*sdkerr.ServiceResponseError); !ok || responseError.StatusCode != http.StatusNotFound {
			return nil, 0, errors.Wrapf(errcode.InternalError, "huawei get repo detail error: %s", err.Error())
		}
	}

	list := make([]*resp.GetDockerTagsResp, 0)

	if repo == nil || repo.NumImages == nil || *repo.NumImages <= 0 {
		return list, 0, nil
	}

	tags, err := c.SwrClient.ListRepositoryTags(&model.ListRepositoryTagsRequest{
		Namespace:   namespace,
		Limit:       &limit,
		Offset:      &offset,
		Repository:  getReq.ProjectName,
		OrderColumn: &orderColumn,
		OrderType:   &orderType,
	})
	if err != nil {
		return nil, 0, errors.Wrapf(errcode.InternalError, "huawei list repo tag error: %s", err.Error())
	}

	for index := range *tags.Body {
		createTime, err := time.Parse(ImageTimeLayout, (*tags.Body)[index].Created)
		if err != nil {
			return nil, 0, err
		}

		updateTime, err := time.Parse(ImageTimeLayout, (*tags.Body)[index].Updated)
		if err != nil {
			return nil, 0, err
		}

		list = append(list, &resp.GetDockerTagsResp{
			ImageID:     (*tags.Body)[index].ImageId,
			Tag:         (*tags.Body)[index].Tag,
			RepoID:      (*tags.Body)[index].RepoId,
			ImageCreate: createTime,
			ImageUpdate: updateTime,
			ImageSize:   (*tags.Body)[index].Size,
			Digest:      (*tags.Body)[index].Digest,
		})
	}

	return list, int(*repo.NumImages), nil
}
