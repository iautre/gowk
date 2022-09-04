package conf

import (
	"fmt"

	"gopkg.in/ini.v1"
)

func Section(name string) *ini.Section {
	return cfg.Section(name)
}

func Get(key string) any {
	return confs[key]
}
func GetString(key string) string {
	return fmt.Sprintf("%v", Get(key))
}
