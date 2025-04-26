package entity

// Internal user.
type InternalUser struct {
	IsActive       int                       `json:"is_active"`
	Nickname       string                    `json:"nickname"`
	Email          string                    `json:"email"`
	DingTalkUserID string                    `json:"dingtalk_user_id"`
	Departments    []*InternalUserDepartment `json:"departments"`
}

// Internal user ding department.
type InternalUserDepartment struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
