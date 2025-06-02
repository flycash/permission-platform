package main

import "gitee.com/flycash/permission-platform/pkg/permission/internal"

func InitClient(addr string) internal.AuthorizedClient {
	c, err := internal.NewGRPCClient(addr)
	if err != nil {
		panic(err)
	}
	return *internal.NewClient(c, "xyz")
}

// PermissionClient 注意：这里的-race和-tags=unit不是必须的，是为了在根目录下执行make ut时client_test.go能够通过，你可以将这两个选项去掉看看效果
//
//go:generate go build -race -tags=unit -mod=readonly -modfile=../../../../go.mod -buildmode=plugin -o ../../testdata/plugins/client.so ./client.go
var PermissionClient = InitClient("192.168.0.1")
