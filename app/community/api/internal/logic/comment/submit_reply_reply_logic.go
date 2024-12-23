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

type SubmitReplyReplyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubmitReplyReplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitReplyReplyLogic {
	return &SubmitReplyReplyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitReplyReplyLogic) SubmitReplyReply(r *http.Request, req *types.ReplyReplyRequest) (resp *types.CommentResponse, err error) {
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

	// 获取客户端IP及位置信息
	ip := utils.GetClientIP(r)
	location, err := l.svcCtx.Ip2Region.SearchByStr(ip)
	if err != nil {
		return nil, err
	}
	location = utils.RemoveZeroAndAdjust(location)

	// 获取浏览器与操作系统信息
	browser, _ := ua.Browser()
	operatingSystem := ua.OS()

	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found in context")
	}

	// 判断作者身份
	var isAuthor int64 = 0
	if uid == req.Author {
		isAuthor = 1
	}

	// XSS过滤
	xssFilterContent := utils.XssFilter(req.Content)
	if xssFilterContent == "" {
		return nil, errors2.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "comment.commentError"))

	}
	commentContent := l.svcCtx.Sensitive.Replace(xssFilterContent, '*')

	tx := l.svcCtx.DB.Begin()
	topicType := constant.CommentTopicType
	commentType := constant.REPLY
	replyReply := &model.ScaCommentReply{
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
		ReplyID:         req.ReplyId,
		ReplyUser:       req.ReplyUser,
		ReplyTo:         req.ReplyTo,
	}
	err = tx.ScaCommentReply.Create(replyReply)
	if err != nil {
		return nil, err
	}
	commentReply := l.svcCtx.DB.ScaCommentReply
	update, err := tx.ScaCommentReply.Where(commentReply.ID.Eq(req.ReplyId)).Update(commentReply.ReplyCount, commentReply.ReplyCount.Add(1))
	if err != nil {
		return nil, err
	}
	if update.RowsAffected == 0 {
		return nil, errors2.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "comment.commentError"))

	}

	// 处理图片
	if len(req.Images) > 0 {
		imagesData, err := utils.ProcessImages(req.Images)
		if err != nil {
			return nil, err
		}
		commentImages := &types.CommentImages{
			UserId:    uid,
			TopicId:   req.TopicId,
			CommentId: replyReply.ID,
			Images:    imagesData,
		}

		newCollection := mongodb.MustNewCollection[types.CommentImages](l.svcCtx.MongoClient, constant.COMMENT_IMAGES)
		_, err = newCollection.Creator().InsertOne(l.ctx, commentImages)
		if err != nil {
			return nil, err
		}
	}

	// 构建响应
	commentResponse := &types.CommentResponse{
		Id:              replyReply.ID,
		Content:         commentContent,
		UserId:          uid,
		TopicId:         replyReply.TopicID,
		Author:          isAuthor,
		Location:        location,
		Browser:         browser,
		OperatingSystem: operatingSystem,
		CreatedTime:     time.Now().Format(constant.TimeFormat),
		ReplyId:         replyReply.ReplyID,
		ReplyUser:       replyReply.ReplyUser,
		ReplyTo:         replyReply.ReplyTo,
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return commentResponse, nil
}
