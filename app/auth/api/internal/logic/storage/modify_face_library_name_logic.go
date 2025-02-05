package storage

import (
	"context"
	"errors"
	"schisandra-album-cloud-microservices/app/aisvc/rpc/pb"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ModifyFaceLibraryNameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewModifyFaceLibraryNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyFaceLibraryNameLogic {
	return &ModifyFaceLibraryNameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ModifyFaceLibraryNameLogic) ModifyFaceLibraryName(req *types.ModifyFaceNameRequestAndResponse) (resp *types.ModifyFaceNameRequestAndResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	faceInfo, err := l.svcCtx.AiSvcRpc.ModifyFaceName(l.ctx, &pb.ModifyFaceNameRequest{UserId: uid, FaceId: req.ID, FaceName: req.FaceName})
	if err != nil {
		return nil, err
	}
	return &types.ModifyFaceNameRequestAndResponse{ID: faceInfo.GetFaceId(), FaceName: faceInfo.GetFaceName()}, nil
}
