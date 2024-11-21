package comment

import (
	"context"
	"errors"

	"schisandra-album-cloud-microservices/app/core/api/common/response"
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
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found in context")
	}
	tx := l.svcCtx.DB.Begin()
	commentLike := l.svcCtx.DB.ScaCommentLike
	resultInfo, err := tx.ScaCommentLike.Where(commentLike.TopicID.Eq(req.TopicId), commentLike.CommentID.Eq(req.CommentId), commentLike.UserID.Eq(uid)).Delete()
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if resultInfo.RowsAffected == 0 {
		_ = tx.Rollback()
		return response.ErrorWithI18n(l.ctx, "comment.CancelLikeError"), nil
	}
	comment := l.svcCtx.DB.ScaCommentReply
	updates, err := tx.ScaCommentReply.Where(comment.TopicID.Eq(req.TopicId), comment.ID.Eq(req.CommentId)).Update(comment.Likes, comment.Likes.Sub(1))
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if updates.RowsAffected == 0 {
		_ = tx.Rollback()
		return response.ErrorWithI18n(l.ctx, "comment.LikeError"), nil
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return response.Success(), nil
}
