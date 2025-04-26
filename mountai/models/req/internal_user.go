package req

type GetInternalUsersReq struct {
	Email string `form:"email" json:"email"`
}
