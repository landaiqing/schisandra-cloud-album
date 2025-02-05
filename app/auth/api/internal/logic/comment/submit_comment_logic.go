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
	"strconv"

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
	commentReply := l.svcCtx.DB.ScaCommentReply
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
	err = commentReply.Create(comment)
	if err != nil {
		return nil, err
	}

	if req.Images != "" {
		imagesData, err := utils.Base64ToBytes(req.Images)
		if err != nil {
			return nil, err
		}
		objectKey := path.Join(
			req.TopicId,
			time.Now().Format("2006/01"), // 按年/月划分目录
			strconv.FormatInt(comment.ID, 10),
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
		info, err := commentReply.Where(commentReply.ID.Eq(comment.ID)).Update(commentReply.ImagePath, objectKey)
		if err != nil || info.RowsAffected == 0 {
			return nil, errors.New("update image path failed")
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
