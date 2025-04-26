package resp

type FavProjectResp struct {
	ID      string             `json:"id" deepcopy:"method:GenerateObjectIDString"`
	User    *UserProfileResp   `json:"user"`
	Project *ProjectDetailResp `json:"project"`
}
