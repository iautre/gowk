package gowk

import (
	"fmt"
	"testing"
)

func TestGo(t *testing.T) {
	Go(func() {
		Fddd("2222")
	})
}

func Fddd(a string) {
	fmt.Println(a)
}
