package comment

import (
	"context"

	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetReplyListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetReplyListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetReplyListLogic {
	return &GetReplyListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetReplyListLogic) GetReplyList(req *types.ReplyListRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}
