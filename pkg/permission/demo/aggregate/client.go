package main

import "gitee.com/flycash/permission-platform/pkg/permission/internal"

func InitClient(_ string) internal.AggregatePermissionClient {
	return internal.AggregatePermissionClient{}
}

//go:generate go build -mod=readonly -modfile=../../../../go.mod -buildmode=plugin -o ../../testdata/plugins/aggregate.so ./client.go
var PermissionClient = InitClient("192.168.0.1")
