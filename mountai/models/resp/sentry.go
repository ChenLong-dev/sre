package resp

type CreateSentryProjectResp struct {
	ProjectSlug string `json:"slug"`
	ProjectID   string `json:"id"`
}

type GetSentryProjectKeyResp struct {
	ProjectDsn SentryProjectDsn `json:"dsn"`
	ProjectID  int              `json:"projectId"`
}

type SentryProjectDsn struct {
	Public   string `json:"public"`
	Security string `json:"security"`
	Secret   string `json:"secret"`
	Csp      string `json:"csp"`
}

type CreateSentryTeamResp struct {
	MemberCount int    `json:"memberCount"`
	TeamSlug    string `json:"slug"`
	TeamID      string `json:"id"`
}

type GetSentryTeamsResp struct {
	TeamSlug string `json:"slug"`
	TeamID   string `json:"id"`
}

type CreateAppSentryResp struct {
	SentryProjectPublicDsn string `json:"sentry_project_public_dsn"`
	SentryProjectSlug      string `json:"sentry_project_slug"`
}

type SentryProjectDetailResp struct {
	Slug  string           `json:"slug"`
	Team  SentryTeamResp   `json:"team"`
	Teams []SentryTeamResp `json:"teams"`
}

type SentryTeamResp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type SentryDingDingResp struct {
	Config []SentryDingConfig `json:"config"`
}

type SentryDingConfig struct {
	Label string `json:"label"`
	Value string `json:"value"`
}
