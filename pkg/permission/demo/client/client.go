package main

import "gitee.com/flycash/permission-platform/pkg/permission/internal"

func InitClient(addr string) internal.AuthorizedClient {
	c, err := internal.NewGRPCClient(addr)
	if err != nil {
		panic(err)
	}
	return *internal.NewClient(c, "xyz")
}

//go:generate go build -mod=readonly -modfile=../../../../go.mod -buildmode=plugin -o ../../testdata/plugins/client.so ./client.go
var PermissionClient = InitClient("192.168.0.1")
