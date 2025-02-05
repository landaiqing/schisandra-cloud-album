package aiservicelogic

import (
	"context"
	"errors"

	"schisandra-album-cloud-microservices/app/aisvc/rpc/internal/svc"
	"schisandra-album-cloud-microservices/app/aisvc/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ModifyFaceTypeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewModifyFaceTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyFaceTypeLogic {
	return &ModifyFaceTypeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ModifyFaceType
func (l *ModifyFaceTypeLogic) ModifyFaceType(in *pb.ModifyFaceTypeRequest) (*pb.ModifyFaceTypeResponse, error) {
	storageFace := l.svcCtx.DB.ScaStorageFace
	faceIds := in.GetFaceId()
	info, err := storageFace.Where(storageFace.ID.In(faceIds...), storageFace.UserID.Eq(in.GetUserId())).Update(storageFace.FaceType, in.GetType())
	if err != nil {
		return nil, err
	}
	if info.RowsAffected == 0 {
		return &pb.ModifyFaceTypeResponse{
			Result: "fail",
		}, errors.New("face not found")
	}
	return &pb.ModifyFaceTypeResponse{
		Result: "success",
	}, nil
}
