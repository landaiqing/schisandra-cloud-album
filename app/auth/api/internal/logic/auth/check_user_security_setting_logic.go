package auth

import (
	"context"
	"errors"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckUserSecuritySettingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCheckUserSecuritySettingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckUserSecuritySettingLogic {
	return &CheckUserSecuritySettingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CheckUserSecuritySettingLogic) CheckUserSecuritySetting() (resp *types.UserSecuritySettingResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	authUser := l.svcCtx.DB.ScaAuthUser
	userSocial := l.svcCtx.DB.ScaAuthUserSocial
	var user model.ScaAuthUser
	err = authUser.Where(authUser.UID.Eq(uid)).Scan(&user)
	if err != nil {
		return nil, err
	}
	// 查询用户社交信息
	var socials []model.ScaAuthUserSocial
	err = userSocial.Where(userSocial.UserID.Eq(uid)).Scan(&socials)
	if err != nil {
		return nil, err
	}
	resp = &types.UserSecuritySettingResponse{
		SetPassword: user.Password != "",
		BindEmail:   user.Email != "",
		BindPhone:   user.Phone != "",
	}
	// 遍历社交信息以设置绑定状态
	for _, social := range socials {
		switch social.Source {
		case "wechat":
			resp.BindWechat = true
		case "qq":
			resp.BindQQ = true
		case "gitee":
			resp.BindGitee = true
		case "github":
			resp.BindGitHub = true
		}
	}
	return resp, nil
}
