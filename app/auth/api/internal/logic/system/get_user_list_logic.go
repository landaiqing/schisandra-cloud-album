package system

import (
	"context"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserListLogic {
	return &GetUserListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserListLogic) GetUserList() (resp *types.UserInfoListResponse, err error) {
	authUser := l.svcCtx.DB.ScaAuthUser
	var userMetaList []types.UserMeta
	err = authUser.Select(
		authUser.ID,
		authUser.UID,
		authUser.Username,
		authUser.Nickname,
		authUser.Email,
		authUser.Phone,
		authUser.Gender,
		authUser.Avatar,
		authUser.Location,
		authUser.Company,
		authUser.Blog,
		authUser.Introduce,
		authUser.Status,
		authUser.CreatedAt,
		authUser.UpdatedAt,
		authUser.DeletedAt,
	).Scan(&userMetaList)
	if err != nil {
		return nil, err
	}
	return &types.UserInfoListResponse{Records: userMetaList}, nil
}
