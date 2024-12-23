package comment

import (
	"context"
	"errors"
	"net/http"
	"schisandra-album-cloud-microservices/app/community/api/internal/svc"
	"schisandra-album-cloud-microservices/app/community/api/internal/types"
	"schisandra-album-cloud-microservices/app/community/api/model/mongodb"
	"schisandra-album-cloud-microservices/app/community/api/model/mysql/model"
	"schisandra-album-cloud-microservices/common/captcha/verify"
	"schisandra-album-cloud-microservices/common/constant"
	errors2 "schisandra-album-cloud-microservices/common/errors"
	"schisandra-album-cloud-microservices/common/i18n"
	"schisandra-album-cloud-microservices/common/utils"
	"time"

	"github.com/mssola/useragent"

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

func (l *SubmitCommentLogic) SubmitComment(r *http.Request, req *types.CommentRequest) (resp *types.CommentResponse, err error) {

	res := verify.VerifySlideCaptcha(l.ctx, l.svcCtx.RedisClient, req.Point, req.Key)
	if !res {
		return nil, errors2.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "captcha.verificationFailure"))
	}
	if len(req.Images) > 3 {
		return nil, errors2.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "comment.tooManyImages"))
	}
	userAgent := r.Header.Get("User-Agent")
	if userAgent == "" {
		return nil, errors2.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "comment.commentError"))
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
	var isAuthor int64 = 0
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found in context")
	}
	if uid == req.Author {
		isAuthor = 1
	}
	xssFilterContent := utils.XssFilter(req.Content)
	if xssFilterContent == "" {
		return nil, errors2.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "comment.commentError"))
	}
	commentContent := l.svcCtx.Sensitive.Replace(xssFilterContent, '*')
	topicType := constant.CommentTopicType
	commentType := constant.COMMENT
	comment := &model.ScaCommentReply{
		Content:         commentContent,
		UserID:          uid,
		TopicID:         req.TopicId,
		TopicType:       topicType,
		CommentType:     commentType,
		Author:          isAuthor,
		CommentIP:       ip,
		Location:        location,
		Browser:         browser,
		OperatingSystem: operatingSystem,
		Agent:           userAgent,
	}
	err = l.svcCtx.DB.ScaCommentReply.Create(comment)
	if err != nil {
		return nil, err
	}

	if len(req.Images) > 0 {
		imagesData, err := utils.ProcessImages(req.Images)
		if err != nil {
			return nil, err
		}
		commentImages := &types.CommentImages{
			UserId:    uid,
			TopicId:   req.TopicId,
			CommentId: comment.ID,
			Images:    imagesData,
		}

		newCollection := mongodb.MustNewCollection[types.CommentImages](l.svcCtx.MongoClient, constant.COMMENT_IMAGES)
		_, err = newCollection.Creator().InsertOne(l.ctx, commentImages)
		if err != nil {
			return nil, err
		}
	}
	commentResponse := &types.CommentResponse{
		Id:              comment.ID,
		Content:         commentContent,
		UserId:          uid,
		TopicId:         req.TopicId,
		Author:          isAuthor,
		Location:        location,
		Browser:         browser,
		OperatingSystem: operatingSystem,
		CreatedTime:     time.Now().Format(constant.TimeFormat),
	}
	return commentResponse, nil
}
