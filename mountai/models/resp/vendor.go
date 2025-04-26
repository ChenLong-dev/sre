package resp

import "time"

type GetDockerTagsResp struct {
	ImageID     string    `json:"imageId"`
	Tag         string    `json:"tag"`
	RepoID      int64     `json:"repoId"`
	ImageUpdate time.Time `json:"imageUpdate"`
	ImageCreate time.Time `json:"imageCreate"`
	ImageSize   int64     `json:"imageSize"`
	Digest      string    `json:"digest"`
}
