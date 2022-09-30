package weapp

import "testing"

func TestGetToken(t *testing.T) {
	token := GetAccessToken()
	t.Log(token)
}
