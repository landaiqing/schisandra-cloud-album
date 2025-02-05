package aiservicelogic

import (
	"context"
	"errors"

	"schisandra-album-cloud-microservices/app/aisvc/rpc/internal/svc"
	"schisandra-album-cloud-microservices/app/aisvc/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ModifyFaceNameLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewModifyFaceNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyFaceNameLogic {
	return &ModifyFaceNameLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ModifyFaceNameLogic) ModifyFaceName(in *pb.ModifyFaceNameRequest) (*pb.ModifyFaceNameResponse, error) {
	storageFace := l.svcCtx.DB.ScaStorageFace
	affected, err := storageFace.Where(storageFace.ID.Eq(in.GetFaceId()), storageFace.UserID.Eq(in.GetUserId())).Update(storageFace.FaceName, in.GetFaceName())
	if err != nil {
		return nil, err
	}
	if affected.RowsAffected == 0 {
		return nil, errors.New("update failed, no rows affected")
	}
	return &pb.ModifyFaceNameResponse{
		FaceId:   in.GetFaceId(),
		FaceName: in.GetFaceName(),
	}, nil
}
