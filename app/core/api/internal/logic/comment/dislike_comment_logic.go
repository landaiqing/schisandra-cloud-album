package comment

import (
	"context"

	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DislikeCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDislikeCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DislikeCommentLogic {
	return &DislikeCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DislikeCommentLogic) DislikeComment(req *types.CommentDisLikeRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}
