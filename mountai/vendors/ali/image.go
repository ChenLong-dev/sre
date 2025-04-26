package ali

import (
	"context"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"

	"rulai/models/req"
	"rulai/models/resp"
)

func (c *Controller) GetRepoTags(ctx context.Context, getReq *req.GetRepoTagsReq) ([]*resp.GetDockerTagsResp, int, error) {
	return nil, 0, errors.Wrap(errcode.InternalError, "not supported yet")
}
