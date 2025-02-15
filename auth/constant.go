package auth

const (
	CONTEXT_APP_KEY = "CONTEXT_APP_KEY"
)

// 用户状态
const (
	DISABLE uint = iota //停用
	ENABLE              //启用
)

// 用户组
const (
	USER_GROUP_ADMIN   = "ADMIN"
	USER_GROUP_DEFAULT = "DEFAULT"
)
