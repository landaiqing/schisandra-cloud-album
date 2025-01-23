package aiservicelogic

import (
	"context"

	"schisandra-album-cloud-microservices/app/aisvc/rpc/internal/svc"
	"schisandra-album-cloud-microservices/app/aisvc/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CaffeClassificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCaffeClassificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CaffeClassificationLogic {
	return &CaffeClassificationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CaffeClassification
func (l *CaffeClassificationLogic) CaffeClassification(in *pb.CaffeClassificationRequest) (*pb.CaffeClassificationResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.CaffeClassificationResponse{}, nil
}
