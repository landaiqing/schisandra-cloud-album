package user

import (
	"context"
	"gorm.io/gorm"
	"net/http"
	"schisandra-album-cloud-microservices/common/captcha/verify"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/errors"
	"schisandra-album-cloud-microservices/common/i18n"
	"schisandra-album-cloud-microservices/common/utils"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminLoginLogic {
	return &AdminLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminLoginLogic) AdminLogin(r *http.Request, req *types.AdminLoginRequest) (resp *types.LoginResponse, err error) {
	captcha := verify.VerifyBasicTextCaptcha(req.Dots, req.Key, l.svcCtx.RedisClient, l.ctx)
	if !captcha {
		return nil, errors.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "captcha.verificationFailure"))
	}
	authUser := l.svcCtx.DB.ScaAuthUser
	permissionRule := l.svcCtx.DB.ScaAuthPermissionRule
	adminUser, err := authUser.
		LeftJoin(permissionRule, authUser.UID.EqCol(permissionRule.V0)).
		Where(authUser.Username.Eq(req.Account), authUser.Password.Eq(req.Password), permissionRule.V1.Eq(constant.Admin)).
		Group(authUser.UID).First()
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if adminUser == nil {
		return nil, errors.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "login.notPermission"))
	}
	if !utils.Verify(adminUser.Password, req.Password) {
		return nil, errors.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "login.invalidPassword"))
	}
	data, err := HandleLoginJWT(adminUser, l.svcCtx, true, r, l.ctx)
	if err != nil {
		return nil, err
	}

	return data, nil
}
