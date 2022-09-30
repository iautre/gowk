package weapp

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/iautre/gowk/log"
)

type Weapp struct{}

type Err struct {
	Errcode int64  `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

type Token struct {
	Err
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	ExpiresTime int64
}

const appId = "wxd713c74ef777d96d"
const secret = "256932c1e4215a71b3d883a3c3fb42cd"
const getAccessTokenUrl = "https://api.weixin.qq.com/cgi-bin/token"

var token *Token = &Token{}
var weapp *Weapp = &Weapp{}

func GetAccessToken() string {
	if token.ExpiresTime <= time.Now().Unix() {
		t, err := weapp.GetAccessToken()
		if err != nil {
			log.Error(context.TODO(), err.Error())
		}
		token = t
	}
	return token.AccessToken
}

// 从微信获取
func (w *Weapp) GetAccessToken() (*Token, error) {
	res, err := http.Get(fmt.Sprintf("%s?grant_type=client_credential&appid=%s&secret=%s", getAccessTokenUrl, appId, secret))
	if err != nil {
		return nil, err
	}
	var t Token
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &t)
	if err != nil {
		return nil, err
	}
	if token.Errcode != 0 {
		return nil, err
	}
	t.ExpiresTime = t.ExpiresIn + time.Now().Unix()
	return &t, nil
}

func (w *Weapp) GetUnlimitedQRCode() (*Token, error) {
	res, err := http.Get(fmt.Sprintf("%s?grant_type=client_credential&appid=%s&secret=%s", getAccessTokenUrl, appId, secret))
	if err != nil {
		return nil, err
	}
	var t Token
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &t)
	if err != nil {
		return nil, err
	}
	if token.Errcode != 0 {
		return nil, err
	}
	t.ExpiresTime = t.ExpiresIn + time.Now().Unix()
	return &t, nil
}
