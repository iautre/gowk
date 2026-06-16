package gowk

// Version 是构建版本，默认 "beta"。
// 构建时通过 ldflags 覆盖：
//
//	go build -ldflags "-X github.com/iautre/gowk.Version=$(git describe --tags --always)" .
var Version = "beta"
