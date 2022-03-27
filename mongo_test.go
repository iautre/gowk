package gowk

import (
	"context"
	"testing"
)

func TestMongo(t *testing.T) {
	t.Log("hello world")
	err := Mongo().Ping(context.TODO(), nil)
	if err != nil {
		t.Fatal(err)
	}
}
