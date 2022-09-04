package gowk

import (
	"context"
	"testing"

	"github.com/iautre/gowk/conf"
)

func TestMongo(t *testing.T) {
	t.Log("hello world")
	err := Mongo().Ping(context.TODO(), nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestConf(t *testing.T) {
	t.Log(conf.Mongo)
}
