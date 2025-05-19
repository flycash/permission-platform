package main

import (
	"context"

	ioc2 "gitee.com/flycash/permission-platform/internal/ioc"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/server/egrpc"
	"go.opentelemetry.io/otel/sdk/trace"

	"gitee.com/flycash/permission-platform/cmd/platform/ioc"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/server"
	"github.com/gotomicro/ego/server/egovernor"
)

func main() {
	// 创建 ego 应用实例
	egoApp := ego.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tp := ioc2.InitZipkinTracer()
	defer func(tp *trace.TracerProvider, ctx context.Context) {
		err := tp.Shutdown(ctx)
		if err != nil {
			elog.Error("Shutdown zipkinTracer", elog.FieldErr(err))
		}
	}(tp, ctx)
	app := ioc.InitApp()

	// 启动服务
	servers := make([]server.Server, 0, len(app.GrpcServers)+1)
	servers = append(servers, egovernor.Load("server.governor").Build())
	servers = append(servers, slice.Map(app.GrpcServers, func(_ int, src *egrpc.Component) server.Server {
		return src
	})...)
	if err := egoApp.Serve(servers...).Run(); err != nil {
		elog.Panic("startup", elog.FieldErr(err))
	}
}
