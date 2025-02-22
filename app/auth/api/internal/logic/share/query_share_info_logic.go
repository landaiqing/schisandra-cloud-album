package share

import (
	"context"
	"errors"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryShareInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryShareInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryShareInfoLogic {
	return &QueryShareInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryShareInfoLogic) QueryShareInfo(req *types.QueryShareInfoRequest) (resp *types.ShareInfoResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}

	storageShare := l.svcCtx.DB.ScaStorageShare
	storageAlbum := l.svcCtx.DB.ScaStorageAlbum
	shareVisit := l.svcCtx.DB.ScaStorageShareVisit
	authUser := l.svcCtx.DB.ScaAuthUser

	var shareInfo types.ShareInfoResponse
	err = storageShare.Select(
		storageShare.ID,
		storageShare.VisitLimit,
		storageShare.InviteCode,
		storageShare.ExpireTime,
		storageShare.CreatedAt,
		storageAlbum.CoverImage,
		storageAlbum.AlbumName,
		storageShare.ImageCount,
		shareVisit.Views.As("visit_count"),
		shareVisit.UserID.Count().As("viewer_count"),
		authUser.Avatar.As("sharer_avatar"),
		authUser.Nickname.As("sharer_name")).
		LeftJoin(storageAlbum, storageShare.AlbumID.EqCol(storageAlbum.ID)).
		Join(shareVisit, storageShare.ID.EqCol(shareVisit.ShareID)).
		LeftJoin(authUser, storageShare.UserID.EqCol(authUser.UID)).
		Where(
			storageShare.InviteCode.Eq(req.InviteCode),
			shareVisit.UserID.Eq(uid)).
		Group(
			storageShare.ID,
			storageShare.VisitLimit,
			storageShare.InviteCode,
			storageShare.ExpireTime,
			storageShare.CreatedAt,
			storageAlbum.CoverImage,
			storageShare.ImageCount,
			storageAlbum.AlbumName,
			shareVisit.Views,
			authUser.Avatar,
			authUser.Nickname).
		Scan(&shareInfo)
	if err != nil {
		return nil, err
	}
	return &shareInfo, nil
}
