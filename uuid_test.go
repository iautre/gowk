package gowk

import "testing"

func TestUUID(t *testing.T) {
	a := UUID()
	t.Log(a)
}
