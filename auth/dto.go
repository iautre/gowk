package auth

type RegisterParams struct {
	Phone string `json:"phone"`
	// Email       string `json:"email"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
type LoginParams struct {
	Account string `json:"account"`
	Code    string `json:"code"`
}

type LoginRes struct {
	Token    string `json:"token"`
	UserId   int64  `json:"userId"`
	Nickname string `json:"nickname"`
}
type UserRes struct {
	Id       int64  `json:"id"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Group    string `json:"group"`
}
