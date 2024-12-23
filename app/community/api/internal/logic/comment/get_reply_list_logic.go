package comment

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"schisandra-album-cloud-microservices/app/community/api/internal/svc"
	types2 "schisandra-album-cloud-microservices/app/community/api/internal/types"
	"schisandra-album-cloud-microservices/app/community/api/model/mongodb"
	constant2 "schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/utils"
	"sync"

	"github.com/chenmingyong0423/go-mongox/v2/builder/query"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetReplyListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	wg     sync.WaitGroup
}

func NewGetReplyListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetReplyListLogic {
	return &GetReplyListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetReplyListLogic) GetReplyList(req *types2.ReplyListRequest) (resp *types2.CommentListPageResponse, err error) {
	// 获取用户ID
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	var replyQueryList []types2.ReplyListQueryResult
	reply := l.svcCtx.DB.ScaCommentReply
	user := l.svcCtx.DB.ScaAuthUser
	commentUser := user.As("comment_user")
	replyUser := user.As("reply_user")

	count, err := reply.Select(
		reply.ID,
		reply.UserID,
		reply.TopicID,
		reply.Content,
		reply.CreatedAt,
		reply.Author,
		reply.ReplyCount,
		reply.Likes,
		reply.Browser,
		reply.OperatingSystem,
		reply.Location,
		reply.ReplyUser,
		reply.ReplyTo,
		reply.ReplyID,
		commentUser.Avatar,
		commentUser.Nickname,
		replyUser.Nickname.As("reply_nickname"),
	).LeftJoin(commentUser, reply.UserID.EqCol(commentUser.UID)).
		LeftJoin(replyUser, reply.ReplyUser.EqCol(replyUser.UID)).
		Where(reply.TopicID.Eq(req.TopicId), reply.ReplyID.Eq(req.CommentId), reply.CommentType.Eq(constant2.REPLY)).
		Order(reply.Likes.Desc(), reply.CreatedAt.Desc()).
		ScanByPage(&replyQueryList, (req.Page-1)*req.Size, req.Size)
	if err != nil {
		return nil, err
	}
	if count == 0 || len(replyQueryList) == 0 {
		return &types2.CommentListPageResponse{
			Total:   count,
			Size:    req.Size,
			Current: req.Page,
		}, nil
	}
	// **************** 获取评论Id和用户Id ************
	commentIds := make([]int64, 0, len(replyQueryList))
	for _, commentList := range replyQueryList {
		commentIds = append(commentIds, commentList.ID)
	}
	l.wg.Add(2)

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
	// ***************获取评论图片 **********
	commentImageMap := make(map[int64][]string)
	go func() {
		defer l.wg.Done()
		newCollection := mongodb.MustNewCollection[types2.CommentImages](l.svcCtx.MongoClient, constant2.COMMENT_IMAGES)
		commentImages, err := newCollection.Finder().
			Filter(query.Eq("topic_id", req.TopicId)).
			Filter(query.In("comment_id", commentIds...)).
			Find(l.ctx)
		if err != nil {
			logx.Error(err)
			return
		}

		for _, image := range commentImages {
			if len(image.Images) == 0 {
				continue
			}
			imagesBase64 := make([]string, len(image.Images))
			for i, img := range image.Images {
				imagesBase64[i] = fmt.Sprintf("data:%s;base64,%s", utils.GetMimeType(img), base64.StdEncoding.EncodeToString(img))
			}
			commentImageMap[image.CommentId] = imagesBase64
		}
	}()
	l.wg.Wait()

	// *************** 组装数据 **********
	result := make([]types2.CommentContent, 0, len(replyQueryList))
	for _, replyData := range replyQueryList {
		commentContent := types2.CommentContent{
			Avatar:          replyData.Avatar,
			NickName:        replyData.Nickname,
			Content:         replyData.Content,
			CreatedTime:     replyData.CreatedAt.Format(constant2.TimeFormat),
			Level:           0,
			Id:              replyData.ID,
			UserId:          replyData.UserID,
			TopicId:         replyData.TopicID,
			IsAuthor:        replyData.Author,
			Likes:           replyData.Likes,
			ReplyCount:      replyData.ReplyCount,
			Location:        replyData.Location,
			Browser:         replyData.Browser,
			OperatingSystem: replyData.OperatingSystem,
			IsLiked:         likeMap[replyData.ID],
			Images:          commentImageMap[replyData.ID],
			ReplyUser:       replyData.ReplyUser,
			ReplyTo:         replyData.ReplyTo,
			ReplyId:         replyData.ReplyId,
			ReplyNickname:   replyData.ReplyNickname,
		}
		result = append(result, commentContent)
	}
	commentListPageResponse := &types2.CommentListPageResponse{
		Total:    count,
		Size:     req.Size,
		Current:  req.Page,
		Comments: result,
	}
	return commentListPageResponse, nil
}
