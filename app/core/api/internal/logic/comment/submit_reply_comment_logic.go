package comment

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/mssola/useragent"

	"schisandra-album-cloud-microservices/app/core/api/common/captcha/verify"
	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/common/utils"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitReplyCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubmitReplyCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitReplyCommentLogic {
	return &SubmitReplyCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitReplyCommentLogic) SubmitReplyComment(r *http.Request, req *types.ReplyCommentRequest) (resp *types.Response, err error) {
	res := verify.VerifySlideCaptcha(l.ctx, l.svcCtx.RedisClient, req.Point, req.Key)
	if !res {
		return response.ErrorWithI18n(l.ctx, "captcha.verificationFailure"), nil
	}
	if len(req.Images) > 3 {
		return response.ErrorWithI18n(l.ctx, "comment.tooManyImages"), nil
	}
	userAgent := r.Header.Get("User-Agent")
	if userAgent == "" {
		return response.ErrorWithI18n(l.ctx, "comment.commentError"), nil
	}
	ua := useragent.New(userAgent)

	ip := utils.GetClientIP(r)
	location, err := l.svcCtx.Ip2Region.SearchByStr(ip)
	if err != nil {
		return nil, err
	}
	location = utils.RemoveZeroAndAdjust(location)

	browser, _ := ua.Browser()
	operatingSystem := ua.OS()
	session, err := l.svcCtx.Session.Get(r, constant.SESSION_KEY)
	if err != nil {
		return nil, err
	}

	uid, ok := session.Values["uid"].(string)
	if !ok {
		return nil, errors.New("uid not found in session")
	}
	isAuthor := 0
	if uid == req.Author {
		isAuthor = 1
	}

	xssFilterContent := utils.XssFilter(req.Content)
	if xssFilterContent == "" {
		return response.ErrorWithI18n(l.ctx, "comment.commentError"), nil
	}
	commentContent := l.svcCtx.Sensitive.Replace(xssFilterContent, '*')

	tx := l.svcCtx.DB.NewSession()
	defer tx.Close()
	if err = tx.Begin(); err != nil {
		return nil, err
	}

	reply := model.ScaCommentReply{
		Content:         commentContent,
		UserId:          uid,
		TopicId:         req.TopicId,
		TopicType:       constant.CommentTopicType,
		CommentType:     constant.COMMENT,
		Author:          isAuthor,
		CommentIp:       ip,
		Location:        location,
		Browser:         browser,
		OperatingSystem: operatingSystem,
		Agent:           userAgent,
		ReplyId:         req.ReplyId,
		ReplyUser:       req.ReplyUser,
	}
	affected, err := tx.Insert(&reply)
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return response.ErrorWithI18n(l.ctx, "comment.commentError"), nil
	}
	update, err := tx.Table(model.ScaCommentReply{}).Where("id = ? and deleted = 0", req.ReplyId).Incr("reply_count", 1).Update(nil)
	if err != nil {
		return nil, err
	}
	if update == 0 {
		return response.ErrorWithI18n(l.ctx, "comment.commentError"), nil
	}

	if len(req.Images) > 0 {
		imagesData, err := utils.ProcessImages(req.Images)
		if err != nil {
			return nil, err
		}

		commentImages := types.CommentImages{
			UserId:    uid,
			TopicId:   req.TopicId,
			CommentId: reply.Id,
			Images:    imagesData,
			CreatedAt: reply.CreatedAt.String(),
		}
		if _, err = l.svcCtx.MongoClient.Collection("comment_images").InsertOne(l.ctx, commentImages); err != nil {
			return nil, err
		}
	}

	commentResponse := types.CommentResponse{
		Id:              reply.Id,
		Content:         commentContent,
		UserId:          uid,
		TopicId:         reply.TopicId,
		Author:          isAuthor,
		Location:        location,
		Browser:         browser,
		OperatingSystem: operatingSystem,
		CreatedTime:     time.Now(),
		ReplyId:         reply.ReplyId,
		ReplyUser:       reply.ReplyUser,
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return response.SuccessWithData(commentResponse), nil
}
