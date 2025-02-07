package storage

import (
	"context"
	"errors"
	"gorm.io/gen/field"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAlbumListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAlbumListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAlbumListLogic {
	return &GetAlbumListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAlbumListLogic) GetAlbumList(req *types.AlbumListRequest) (resp *types.AlbumListResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	var orderConditions []field.Expr
	storageAlbum := l.svcCtx.DB.ScaStorageAlbum
	if req.Sort {
		orderConditions = append(orderConditions, storageAlbum.CreatedAt.Desc())
	} else {
		orderConditions = append(orderConditions, storageAlbum.AlbumName.Desc())
	}
	albums, err := storageAlbum.Where(storageAlbum.UserID.Eq(uid), storageAlbum.AlbumType.Eq(req.Type)).Order(orderConditions...).Find()
	if err != nil {
		return nil, err
	}
	var albumList []types.Album
	for _, album := range albums {
		albumList = append(albumList, types.Album{
			ID:         album.ID,
			Name:       album.AlbumName,
			Type:       album.AlbumType,
			CoverImage: album.CoverImage,
			CreatedAt:  album.CreatedAt.Format("2006-01-02"),
		})
	}
	return &types.AlbumListResponse{
		Albums: albumList,
	}, nil
}
