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

type SubmitCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubmitCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitCommentLogic {
	return &SubmitCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitCommentLogic) SubmitComment(r *http.Request, req *types.CommentRequest) (resp *types.Response, err error) {

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
	isAuthor := 0
	session, err := l.svcCtx.Session.Get(r, constant.SESSION_KEY)
	if err == nil {
		return nil, err
	}
	uid, ok := session.Values["uid"].(string)
	if !ok {
		return nil, errors.New("uid not found in session")
	}
	if uid == req.Author {
		isAuthor = 1
	}
	xssFilterContent := utils.XssFilter(req.Content)
	if xssFilterContent == "" {
		return response.ErrorWithI18n(l.ctx, "comment.commentError"), nil
	}
	commentContent := l.svcCtx.Sensitive.Replace(xssFilterContent, '*')
	comment := model.ScaCommentReply{
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
	}
	affected, err := l.svcCtx.DB.InsertOne(&comment)
	if err != nil {
		return nil, err
	}
	if affected == 0 {
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
			CommentId: comment.Id,
			Images:    imagesData,
			CreatedAt: comment.CreatedAt.String(),
		}
		if _, err = l.svcCtx.MongoClient.Collection("comment_images").InsertOne(l.ctx, commentImages); err != nil {
			return nil, err
		}
	}
	commentResponse := types.CommentResponse{
		Id:              comment.Id,
		Content:         commentContent,
		UserId:          uid,
		TopicId:         req.TopicId,
		Author:          isAuthor,
		Location:        location,
		Browser:         browser,
		OperatingSystem: operatingSystem,
		CreatedTime:     time.Now(),
	}
	return response.SuccessWithData(commentResponse), nil
}
