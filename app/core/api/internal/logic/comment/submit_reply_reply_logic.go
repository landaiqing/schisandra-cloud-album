package comment

import (
	"context"

	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitReplyReplyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubmitReplyReplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitReplyReplyLogic {
	return &SubmitReplyReplyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitReplyReplyLogic) SubmitReplyReply(req *types.ReplyReplyRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}
