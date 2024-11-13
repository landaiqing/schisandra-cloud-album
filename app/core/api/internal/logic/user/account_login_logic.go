package user

import (
	"context"

	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"
	"schisandra-album-cloud-microservices/app/core/api/repository/mongodb/collection"
	"schisandra-album-cloud-microservices/app/core/api/repository/mongodb/model"
	"schisandra-album-cloud-microservices/common/i18n"

	"github.com/zeromicro/go-zero/core/logx"
)

type AccountLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAccountLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AccountLoginLogic {
	return &AccountLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AccountLoginLogic) AccountLogin(req *types.AccountLoginRequest) (resp *types.LoginResponse, err error) {
	// todo: add your logic here and delete this line
	i18n.IsHasI18n(l.ctx)
	text := i18n.FormatText(l.ctx, "user.name", "landaiqing")
	collection.MustNewCollection[model.CommentImage](l.svcCtx, "comment_image")

	return &types.LoginResponse{
		AccessToken: text,
	}, nil
}
