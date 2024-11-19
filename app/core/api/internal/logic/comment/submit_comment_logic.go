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
	session, wrong := l.svcCtx.Session.Get(r, constant.SESSION_KEY)
	if wrong == nil {
		return nil, wrong
	}
	uid := session.Values["uid"].(string)
	if uid == req.Author {
		isAuthor = 1
	}
	xssFilterContent := utils.XssFilter(req.Content)
	if xssFilterContent == "" {
		return response.ErrorWithI18n(l.ctx, "comment.commentError"), nil
	}
	commentContent := l.svcCtx.Sensitive.Replace(xssFilterContent, '*')
	comment, err := l.svcCtx.MySQLClient.ScaCommentReply.Create().
		SetContent(commentContent).
		SetUserID(uid).
		SetTopicID(req.TopicId).
		SetCommentType(constant.CommentTopicType).
		SetCommentType(constant.COMMENT).
		SetAuthor(isAuthor).
		SetCommentIP(ip).
		SetLocation(location).
		SetBrowser(browser).
		SetOperatingSystem(operatingSystem).
		SetAgent(userAgent).Save(l.ctx)
	if err != nil {
		return nil, err
	}
	if len(req.Images) > 0 {
		imagesDataCh := make(chan [][]byte)
		go func() {
			imagesData, err := utils.ProcessImages(req.Images)
			if err != nil {
				imagesDataCh <- nil
				return
			}
			imagesDataCh <- imagesData
		}()
		imagesData := <-imagesDataCh
		if imagesData == nil {
			return nil, errors.New("process images failed")
		}
		commentImages := types.CommentImages{
			UserId:    uid,
			TopicId:   req.TopicId,
			CommentId: comment.ID,
			Images:    imagesData,
			CreatedAt: comment.CreatedAt.String(),
		}
		if _, err = l.svcCtx.MongoClient.Collection("comment_images").InsertOne(l.ctx, commentImages); err != nil {
			return nil, err
		}
	}
	commentResponse := types.CommentResponse{
		Id:              comment.ID,
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
