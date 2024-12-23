package comment

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"net/http"
	"schisandra-album-cloud-microservices/app/community/api/internal/svc"
	"schisandra-album-cloud-microservices/app/community/api/internal/types"
	errors2 "schisandra-album-cloud-microservices/common/errors"
	"schisandra-album-cloud-microservices/common/i18n"
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

func (l *DislikeCommentLogic) DislikeComment(req *types.CommentDisLikeRequest) (err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return errors.New("user_id not found")
	}
	tx := l.svcCtx.DB.Begin()
	commentLike := l.svcCtx.DB.ScaCommentLike
	resultInfo, err := tx.ScaCommentLike.Where(commentLike.TopicID.Eq(req.TopicId), commentLike.CommentID.Eq(req.CommentId), commentLike.UserID.Eq(uid)).Delete()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if resultInfo.RowsAffected == 0 {
		_ = tx.Rollback()
		return errors2.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "comment.LikeError"))
	}
	comment := l.svcCtx.DB.ScaCommentReply
	updates, err := tx.ScaCommentReply.Where(comment.TopicID.Eq(req.TopicId), comment.ID.Eq(req.CommentId)).Update(comment.Likes, comment.Likes.Sub(1))
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if updates.RowsAffected == 0 {
		_ = tx.Rollback()
		return errors2.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "comment.LikeError"))
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
