package comment

import (
	"context"
	"errors"
	"time"

	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type LikeCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLikeCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeCommentLogic {
	return &LikeCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LikeCommentLogic) LikeComment(req *types.CommentLikeRequest) (resp *types.Response, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found in context")
	}
	tx := l.svcCtx.DB.Begin()
	commentLike := &model.ScaCommentLike{
		CommentID: req.CommentId,
		TopicID:   req.TopicId,
		UserID:    uid,
		LikeTime:  time.Now(),
	}
	err = tx.ScaCommentLike.Create(commentLike)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	comment := l.svcCtx.DB.ScaCommentReply
	updates, err := tx.ScaCommentReply.Where(comment.TopicID.Eq(req.TopicId), comment.ID.Eq(req.CommentId)).Update(comment.Likes, comment.Likes.Add(1))
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
