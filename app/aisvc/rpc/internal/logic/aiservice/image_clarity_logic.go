package aiservicelogic

import (
	"bytes"
	"context"
	"image"

	"schisandra-album-cloud-microservices/app/aisvc/rpc/internal/svc"
	"schisandra-album-cloud-microservices/app/aisvc/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ImageClarityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewImageClarityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ImageClarityLogic {
	return &ImageClarityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ImageClarity 图像清晰度检测
func (l *ImageClarityLogic) ImageClarity(in *pb.ImageClarityRequest) (*pb.ImageClarityResponse, error) {
	img, _, err := image.Decode(bytes.NewReader(in.Image))
	if err != nil {
		return nil, err
	}
	check, err := l.svcCtx.ClarityDetector.ClarityCheck(img)
	if err != nil {
		return nil, err
	}
	return &pb.ImageClarityResponse{
		IsBlurred: check,
	}, nil
}
