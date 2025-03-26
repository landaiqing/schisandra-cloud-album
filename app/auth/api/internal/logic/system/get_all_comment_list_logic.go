package system

import (
	"context"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAllCommentListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAllCommentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAllCommentListLogic {
	return &GetAllCommentListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAllCommentListLogic) GetAllCommentList() (resp *types.AllCommentListResponse, err error) {
	commentReply := l.svcCtx.DB.ScaCommentReply
	var records []types.CommentReplyMeta
	err = commentReply.Scan(&records)
	if err != nil {
		return nil, err
	}

	return &types.AllCommentListResponse{
		Records: records,
	}, nil
}
