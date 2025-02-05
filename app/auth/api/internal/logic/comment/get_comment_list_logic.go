package comment

import (
	"context"
	"errors"
	"net/url"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"
	"schisandra-album-cloud-microservices/common/constant"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gen/field"
)

type GetCommentListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	wg     sync.WaitGroup
}

func NewGetCommentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommentListLogic {
	return &GetCommentListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCommentListLogic) GetCommentList(req *types.CommentListRequest) (resp *types.CommentListPageResponse, err error) {
	// 获取用户ID
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	var commentQueryList []types.CommentListQueryResult
	comment := l.svcCtx.DB.ScaCommentReply
	user := l.svcCtx.DB.ScaAuthUser
	var orderConditions []field.Expr

	if req.IsHot {
		orderConditions = append(orderConditions, comment.Likes.Desc(), comment.ReplyCount.Desc())
	} else {
		orderConditions = append(orderConditions, comment.CreatedAt.Desc())
	}
	count, err := comment.Select(
		comment.ID,
		comment.UserID,
		comment.TopicID,
		comment.Content,
		comment.CreatedAt,
		comment.Author,
		comment.Likes,
		comment.ReplyCount,
		comment.Browser,
		comment.OperatingSystem,
		comment.Location,
		comment.ImagePath,
		user.Avatar,
		user.Nickname,
	).LeftJoin(user, comment.UserID.EqCol(user.UID)).
		Where(comment.TopicID.Eq(req.TopicId), comment.CommentType.Eq(constant.COMMENT)).
		Order(orderConditions...).
		ScanByPage(&commentQueryList, (req.Page-1)*req.Size, req.Size)
	if err != nil {
		return nil, err
	}
	if count == 0 || len(commentQueryList) == 0 {
		return &types.CommentListPageResponse{
			Total:    count,
			Size:     req.Size,
			Current:  req.Page,
			Comments: []types.CommentContent{},
		}, nil
	}
	// **************** 获取评论Id和用户Id ************
	commentIds := make([]int64, 0, len(commentQueryList))
	for _, commentList := range commentQueryList {
		commentIds = append(commentIds, commentList.ID)
	}
	l.wg.Add(1)

	// *************** 获取评论点赞状态 **********
	likeMap := make(map[int64]bool)
	go func() {
		defer l.wg.Done()
		commentLike := l.svcCtx.DB.ScaCommentLike
		likeList, err := commentLike.Where(
			commentLike.TopicID.Eq(req.TopicId),
			commentLike.UserID.Eq(uid),
			commentLike.CommentID.In(commentIds...)).
			Find()
		if err != nil {
			logx.Error(err)
			return
		}
		for _, like := range likeList {
			likeMap[like.CommentID] = true
		}
	}()

	l.wg.Wait()

	// *************** 组装数据 **********
	result := make([]types.CommentContent, 0, len(commentQueryList))
	for _, commentData := range commentQueryList {
		var imagePath string
		if commentData.ImagePath != "" {
			reqParams := make(url.Values)
			presignedURL, err := l.svcCtx.MinioClient.PresignedGetObject(l.ctx, constant.CommentImagesBucketName, commentData.ImagePath, time.Hour*24, reqParams)
			if err != nil {
				logx.Error(err)
				continue
			}
			imagePath = presignedURL.String()
		}
		commentContent := types.CommentContent{
			Avatar:          commentData.Avatar,
			NickName:        commentData.Nickname,
			Content:         commentData.Content,
			CreatedTime:     commentData.CreatedAt.Format(constant.TimeFormat),
			Level:           0,
			Id:              commentData.ID,
			UserId:          commentData.UserID,
			TopicId:         commentData.TopicID,
			IsAuthor:        commentData.Author,
			Likes:           commentData.Likes,
			ReplyCount:      commentData.ReplyCount,
			Location:        commentData.Location,
			Browser:         commentData.Browser,
			OperatingSystem: commentData.OperatingSystem,
			IsLiked:         likeMap[commentData.ID],
			Images:          imagePath,
		}
		result = append(result, commentContent)
	}
	commentListPageResponse := &types.CommentListPageResponse{
		Total:    count,
		Size:     req.Size,
		Current:  req.Page,
		Comments: result,
	}
	return commentListPageResponse, nil
}
