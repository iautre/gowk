package gowk

import (
	"crypto/tls"
	"net/http"
	"reflect"
)

func Struct2Map(obj any) (data map[string]any, err error) {
	data = make(map[string]any)
	objT := reflect.TypeOf(obj)
	objV := reflect.ValueOf(obj)
	for i := 0; i < objT.NumField(); i++ {
		data[objT.Field(i).Name] = objV.Field(i).Interface()
	}
	err = nil
	return
}

func HttpClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{Transport: tr}
}
