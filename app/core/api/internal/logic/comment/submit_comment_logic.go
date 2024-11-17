package comment

import (
	"context"

	"schisandra-album-cloud-microservices/app/core/api/common/captcha/verify"
	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubmitCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitCommentLogic {
	return &SubmitCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitCommentLogic) SubmitComment(req *types.CommentRequest) (resp *types.Response, err error) {

	res := verify.VerifySlideCaptcha(l.ctx, l.svcCtx.RedisClient, req.Point, req.Key)
	if !res {
		return response.ErrorWithI18n(l.ctx, "captcha.verificationFailure"), nil
	}

	return
}
