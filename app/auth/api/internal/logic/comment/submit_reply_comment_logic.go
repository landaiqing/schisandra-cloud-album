package comment

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/ccpwcn/kgo"
	"github.com/minio/minio-go/v7"
	"net/http"
	"path"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/model"
	"schisandra-album-cloud-microservices/common/captcha/verify"
	"schisandra-album-cloud-microservices/common/constant"
	errors2 "schisandra-album-cloud-microservices/common/errors"
	"schisandra-album-cloud-microservices/common/i18n"
	"schisandra-album-cloud-microservices/common/utils"
	"strconv"
	"time"

	"github.com/mssola/useragent"

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

func (l *SubmitReplyCommentLogic) SubmitReplyComment(r *http.Request, req *types.ReplyCommentRequest) (resp *types.CommentResponse, err error) {
	res := verify.VerifySlideCaptcha(l.ctx, l.svcCtx.RedisClient, req.Point, req.Key)
	if !res {
		return nil, errors2.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "captcha.verificationFailure"))
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
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found in context")
	}
	var isAuthor int64 = 0
	if uid == req.Author {
		isAuthor = 1
	}

	xssFilterContent := utils.XssFilter(req.Content)
	if xssFilterContent == "" {
		return nil, errors2.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "comment.commentError"))

	}
	commentContent := l.svcCtx.Sensitive.Replace(xssFilterContent, '*')

	tx := l.svcCtx.DB.Begin()
	topicType := constant.CommentTopicType
	commentType := constant.REPLY
	reply := &model.ScaCommentReply{
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
	}
	err = tx.ScaCommentReply.Create(reply)
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

	if req.Images != "" {
		imagesData, err := utils.Base64ToBytes(req.Images)
		if err != nil {
			return nil, err
		}
		objectKey := path.Join(
			req.TopicId,
			time.Now().Format("2006/01"), // 按年/月划分目录
			strconv.FormatInt(reply.ID, 10),
			fmt.Sprintf("%s_%s.jpg", uid, kgo.SimpleUuid()),
		)

		exists, err := l.svcCtx.MinioClient.BucketExists(l.ctx, constant.CommentImagesBucketName)
		if err != nil || !exists {
			err = l.svcCtx.MinioClient.MakeBucket(l.ctx, constant.CommentImagesBucketName, minio.MakeBucketOptions{Region: "us-east-1", ObjectLocking: true})
			if err != nil {
				logx.Errorf("Failed to create MinIO bucket: %v", err)
				return nil, err
			}
		}

		// 上传到MinIO
		_, err = l.svcCtx.MinioClient.PutObject(
			l.ctx,
			constant.CommentImagesBucketName,
			objectKey,
			bytes.NewReader(imagesData),
			int64(len(imagesData)),
			minio.PutObjectOptions{
				ContentType: "image/jpeg",
			},
		)
		if err != nil {
			logx.Errorf("Failed to upload image to MinIO: %v", err)
			return nil, err
		}
		info, err := commentReply.Where(commentReply.ID.Eq(reply.ID)).Update(commentReply.ImagePath, objectKey)
		if err != nil || info.RowsAffected == 0 {
			return nil, errors.New("update image path failed")
		}

	}

	commentResponse := &types.CommentResponse{
		Id:              reply.ID,
		Content:         commentContent,
		UserId:          uid,
		TopicId:         reply.TopicID,
		Author:          isAuthor,
		Location:        location,
		Browser:         browser,
		OperatingSystem: operatingSystem,
		CreatedTime:     time.Now().Format(constant.TimeFormat),
		ReplyId:         reply.ReplyID,
		ReplyUser:       reply.ReplyUser,
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return commentResponse, nil
}
