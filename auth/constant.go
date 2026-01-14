package auth

const (
	// User context keys
	ContextUserID            = "user_id"
	ContextIsAdmin           = "is_admin"
	ContextTargetUserID      = "target_user_id"
	ContextAuthenticatedUser = "authenticated_user"

	// OAuth2 context keys
	ContextOAuth2Token = "oauth2_token"

	// Error messages
	ErrAuthRequired       = "Authentication required"
	ErrAdminRequired      = "Admin access required"
	ErrInvalidCredentials = "Invalid credentials"
	ErrUserNotFound       = "User not found"
	ErrInvalidToken       = "Invalid or expired token"
	ErrInsufficientScope  = "Insufficient scope"
	ErrAccessDenied       = "Access denied"
	ErrInvalidParameter   = "Invalid parameter"
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
