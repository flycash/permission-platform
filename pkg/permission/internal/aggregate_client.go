package internal

import (
	"context"
	"time"

	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
	"google.golang.org/grpc"
)

const (
	defaultSleepTime = 100 * time.Millisecond
	defaultTimeout   = 5 * time.Second
)

type AggregatePermissionClient struct {
	requestsCh      chan *AggregateRequest
	batchClient     permissionv1.BatchPermissionServiceClient
	logger          *elog.Component
	batchSize       int
	permissionToken string
}

type AggregateRequest struct {
	Req    *permissionv1.CheckPermissionRequest
	RespCh chan *permissionv1.CheckPermissionResponse
	ErrCh  chan error
}

func (l *AggregatePermissionClient) Name() string {
	return "AggregateClient"
}

func (l *AggregatePermissionClient) CheckPermission(ctx context.Context, in *permissionv1.CheckPermissionRequest, _ ...grpc.CallOption) (*permissionv1.CheckPermissionResponse, error) {
	req := &AggregateRequest{
		Req:    in,
		RespCh: make(chan *permissionv1.CheckPermissionResponse, 1),
		ErrCh:  make(chan error, 1),
	}
	l.requestsCh <- req
	select {
	case resp := <-req.RespCh:
		return resp, nil
	case err := <-req.ErrCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (l *AggregatePermissionClient) StartBatch(ctx context.Context) {
	for {
		l.oneLoop(ctx)
	}
}

func (l *AggregatePermissionClient) oneLoop(ctx context.Context) {
	reqs := make([]*AggregateRequest, 0, l.batchSize)
	for {
		select {
		case req := <-l.requestsCh:
			reqs = append(reqs, req)
			if len(reqs) < l.batchSize {
				continue
			}
		case <-ctx.Done():
			return
		default:
			if len(reqs) == 0 {
				time.Sleep(defaultSleepTime)
				continue
			}
			ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
			l.batchSend(ctx, reqs)
			cancel()
		}
	}
}

// 发送
func (l *AggregatePermissionClient) batchSend(ctx context.Context, reqs []*AggregateRequest) {
	checkReqs := slice.Map(reqs, func(_ int, src *AggregateRequest) *permissionv1.CheckPermissionRequest {
		return src.Req
	})
	pCtx := context.WithValue(ctx, "Authorization", l.permissionToken)
	resp, err := l.batchClient.BatchCheckPermission(pCtx, &permissionv1.BatchCheckPermissionRequest{
		Requests: checkReqs,
	})
	if err != nil {
		for _, req := range reqs {
			req.ErrCh <- err
		}
		l.logger.Error("批量权限校验失败",
			elog.FieldErr(err),
			elog.Any("reqs", reqs))
		return
	}
	for idx := range reqs {
		req := reqs[idx]
		req.RespCh <- &permissionv1.CheckPermissionResponse{
			Allowed: resp.Allowed[idx],
		}
	}
}
