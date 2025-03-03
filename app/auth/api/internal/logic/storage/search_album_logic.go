package storage

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"
)

type SearchAlbumLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchAlbumLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchAlbumLogic {
	return &SearchAlbumLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchAlbumLogic) SearchAlbum(req *types.SearchAlbumRequest) (resp *types.SearchAlbumResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	storageAlbum := l.svcCtx.DB.ScaStorageAlbum
	storageAlbums, err := storageAlbum.Where(storageAlbum.UserID.Eq(uid), storageAlbum.AlbumName.Like("%"+req.Keyword+"%")).Find()
	if err != nil {
		return nil, err
	}
	if len(storageAlbums) == 0 {
		return nil, nil
	}
	var albums []types.Album
	for _, album := range storageAlbums {
		albums = append(albums, types.Album{
			ID:         album.ID,
			Name:       album.AlbumName,
			Type:       album.AlbumType,
			CoverImage: album.CoverImage,
			CreatedAt:  album.CreatedAt.Format("2006-01-02"),
		})
	}
	return &types.SearchAlbumResponse{
		Albums: albums,
	}, nil
}
