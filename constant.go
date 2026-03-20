package gowk

const (
	AUID          = "auid"
	UID           = "uid"
	APPKEY        = "AppKey"
	AUTHORIZATION = "Authorization"
	ATOKEN        = "Atoken"
)

// 分隔符（不可声明为 const，因为 string([]byte{…}) 不是编译期常量）
var (
	SEP1 = string([]byte{0x01})
	SEP2 = string([]byte{0x02})
)

const WEB_SOCKET_CLIENT_NAME = "web_socket_client_name"
