package req

type CreateSentryProjectReq struct {
	ProjectName string `json:"project_name"`
	ProjectSlug string `json:"project_slug"`
	TeamSlug    string `json:"team_slug"`
}

type EnableSentryDingDingReq struct {
	ProjectSlug string `json:"project_slug"`
}

type UpdateSentryDingDingReq struct {
	ProjectSlug string `json:"project_slug"`
	AccessToken string `json:"access_token"`
}

type GetSentryDingDingReq struct {
	ProjectSlug string `json:"project_slug"`
}

type GetSentryProjectReq struct {
	ProjectSlug string `json:"project_slug"`
}

type AddSentryProjectTeamReq struct {
	ProjectSlug string `json:"project_slug"`
	TeamSlug    string `json:"team_slug"`
}

type RemoveSentryProjectTeamReq struct {
	ProjectSlug string `json:"project_slug"`
	TeamSlug    string `json:"team_slug"`
}

type UpdateSentryProjectTeamReq struct {
	ProjectSlug string `json:"project_slug"`
	TeamSlug    string `json:"team_slug"`
}

type CreateSentryTeamReq struct {
	TeamName string `json:"team_name"`
	TeamSlug string `json:"team_slug"`
}

type GetSentryProjectKeyReq struct {
	ProjectSlug string `json:"project_slug"`
}

type DeleteSentryProjectReq struct {
	ProjectSlug string `json:"project_slug"`
}
