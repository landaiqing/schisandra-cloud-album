package comment

import (
	"context"
	"errors"
	"schisandra-album-cloud-microservices/app/community/api/internal/svc"
	"schisandra-album-cloud-microservices/app/community/api/internal/types"
	"schisandra-album-cloud-microservices/app/community/api/model/mysql/model"
	"schisandra-album-cloud-microservices/common/i18n"
	"time"

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

func (l *LikeCommentLogic) LikeComment(req *types.CommentLikeRequest) (err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return errors.New("user_id not found")
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
		return err
	}
	comment := l.svcCtx.DB.ScaCommentReply
	updates, err := tx.ScaCommentReply.Where(comment.TopicID.Eq(req.TopicId), comment.ID.Eq(req.CommentId)).Update(comment.Likes, comment.Likes.Add(1))
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if updates.RowsAffected == 0 {
		_ = tx.Rollback()
		return errors.New(i18n.FormatText(l.ctx, "comment.LikeError"))
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
