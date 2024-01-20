package conf

import (
	"fmt"
	"testing"
)

type Oss struct {
	Endpoint string `json:"endpoint"`
	Bucket   string `json:"bucket"`
}

func TestConf(t *testing.T) {
	Init("conf.toml")
	a := Gets[Oss]("oss")
	fmt.Println(a)
}
