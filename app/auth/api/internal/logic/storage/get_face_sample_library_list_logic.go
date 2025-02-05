package storage

import (
	"context"
	"errors"
	"schisandra-album-cloud-microservices/app/aisvc/rpc/pb"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFaceSampleLibraryListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFaceSampleLibraryListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFaceSampleLibraryListLogic {
	return &GetFaceSampleLibraryListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFaceSampleLibraryListLogic) GetFaceSampleLibraryList(req *types.FaceSampleLibraryListRequest) (resp *types.FaceSampleLibraryListResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	faceLibrary, err := l.svcCtx.AiSvcRpc.QueryFaceLibrary(l.ctx, &pb.QueryFaceLibraryRequest{UserId: uid, Type: req.Type})
	if err != nil {
		return nil, err
	}
	var faceSampleLibraries []types.FaceSampleLibrary
	for _, face := range faceLibrary.GetFaces() {
		faceSampleLibraries = append(faceSampleLibraries, types.FaceSampleLibrary{
			ID:        face.GetId(),
			FaceName:  face.GetFaceName(),
			FaceImage: face.GetFaceImage(),
		})
	}
	return &types.FaceSampleLibraryListResponse{
		Faces: faceSampleLibraries,
	}, nil
}
