package comment

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"
)

type GetCommentListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCommentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommentListLogic {
	return &GetCommentListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCommentListLogic) GetCommentList(req *types.CommentListRequest) (resp *types.Response, err error) {
	// 获取用户ID
	// uid, ok := l.ctx.Value("user_id").(string)
	// if !ok {
	// 	return nil, errors.New("user_id not found in context")
	// }

	// 查询评论列表
	return
}
