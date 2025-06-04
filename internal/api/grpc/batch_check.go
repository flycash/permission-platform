package grpc

import (
	"context"

	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"gitee.com/flycash/permission-platform/internal/api/grpc/interceptor/auth"
	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/service/hybrid"
	"github.com/ecodeclub/ekit/list"
	"golang.org/x/sync/errgroup"
)

type BatchPermissionServer struct {
	permissionv1.UnimplementedBatchPermissionServiceServer
	permissionSvc hybrid.PermissionService
}

func (b *BatchPermissionServer) BatchCheckPermission(ctx context.Context, request *permissionv1.BatchCheckPermissionRequest) (*permissionv1.BatchCheckPermissionResponse, error) {
	var eg errgroup.Group
	reqs := request.GetRequests()
	bizID, err := auth.GetBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	res := &list.ConcurrentList[bool]{
		List: list.NewArrayListOf[bool](make([]bool, len(reqs))),
	}
	for idx := range reqs {
		req := reqs[idx]
		eg.Go(func() error {
			allow, eerr := b.permissionSvc.Check(ctx, bizID, req.Uid, domain.Resource{
				BizID: bizID,
				Type:  req.Permission.ResourceType,
				Key:   req.Permission.ResourceKey,
			}, req.Permission.Actions,
				domain.Attributes{
					Subject:     b.toSubAttrs(req.SubjectAttributes),
					Resource:    b.toSubAttrs(req.ResourceAttributes),
					Environment: b.toSubAttrs(req.EnvironmentAttributes),
				})
			_ = res.Set(idx, allow)
			return eerr
		})
	}
	err = eg.Wait()
	if err != nil {
		return nil, err
	}
	return &permissionv1.BatchCheckPermissionResponse{
		Allowed: res.AsSlice(),
	}, nil
}

func (b *BatchPermissionServer) toSubAttrs(req map[string]string) domain.SubAttrs {
	var res domain.SubAttrs
	for defName, defValue := range req {
		res = res.SetKv(defName, defValue)
	}
	return res
}
