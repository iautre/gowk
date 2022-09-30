package auth

import (
	"fmt"
)

type AuthType string

var defaultAuth *Auth = &Auth{}

const (
	Brear, Api AuthType = "brear", "api"
)
const (
	A_ID, U_ID string = "A_ID", "U_ID"
)

func (a AuthType) toString() string {
	return fmt.Sprint(a)
}

type Config struct {
	ApiKeyName string
	JwtSecret  []byte
}

type AuthInterface interface {
	CheckByUsernameAndPassword(username string, password string) (string, bool)
	CheckApiKey(key string) (string, bool)
}

type Auth struct {
	Config      *Config
	Type        []AuthType
	HandlerFunc AuthInterface
}

func New(conf *Config, hf AuthInterface, aType ...AuthType) *Auth {
	return &Auth{
		Config:      conf,
		HandlerFunc: hf,
		Type:        aType,
	}
}

func Init(conf *Config, hf AuthInterface, aType ...AuthType) *Auth {
	defaultAuth = New(conf, hf, aType...)
	return defaultAuth
}
