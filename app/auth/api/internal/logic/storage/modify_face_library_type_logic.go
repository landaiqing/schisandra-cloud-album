package storage

import (
	"context"
	"errors"
	"schisandra-album-cloud-microservices/app/aisvc/rpc/pb"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ModifyFaceLibraryTypeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewModifyFaceLibraryTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyFaceLibraryTypeLogic {
	return &ModifyFaceLibraryTypeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ModifyFaceLibraryTypeLogic) ModifyFaceLibraryType(req *types.ModifyFaceTypeRequest) (resp *types.ModifyFaceTypeResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	faceInfo, err := l.svcCtx.AiSvcRpc.ModifyFaceType(l.ctx, &pb.ModifyFaceTypeRequest{UserId: uid, FaceId: req.IDs, Type: req.FaceType})
	if err != nil {
		return nil, err
	}
	storageInfo := l.svcCtx.DB.ScaStorageInfo
	resultInfo, err := storageInfo.Where(storageInfo.FaceID.In(req.IDs...)).Update(storageInfo.ImgShow, req.FaceType)
	if err != nil {
		return nil, err
	}
	if resultInfo.RowsAffected != int64(len(req.IDs)) {
		return nil, errors.New("update failed")
	}
	return &types.ModifyFaceTypeResponse{Result: faceInfo.Result}, nil
}
