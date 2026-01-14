package auth

import (
	"context"
	"testing"
)

func TestOAuth2ClientService_CreateOAuth2Client(t *testing.T) {
	ctx := context.Background()
	var oc OAuth2ClientService
	o, err := oc.CreateOAuth2Client(ctx, &OAuth2ClientCreateParams{
		Name:         "blinko",
		RedirectURIs: []string{"http://192.168.199.100:1111/api/auth/callback/1212121"},
		Scopes:       []string{""},
		GrantTypes:   []string{"code"},
	})
	t.Error(err)
	t.Log(o)
}
