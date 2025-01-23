package aiservicelogic

import (
	"context"
	"fmt"
	"gocv.io/x/gocv"
	"image"
	"schisandra-album-cloud-microservices/app/aisvc/rpc/internal/svc"
	"schisandra-album-cloud-microservices/app/aisvc/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type TfClassificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTfClassificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TfClassificationLogic {
	return &TfClassificationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// TfClassification is a server endpoint to classify an image using TensorFlow.
func (l *TfClassificationLogic) TfClassification(in *pb.TfClassificationRequest) (*pb.TfClassificationResponse, error) {
	className, source, err := l.ClassifyImage(in.GetImage())
	if err != nil {
		return nil, err
	}
	return &pb.TfClassificationResponse{
		Score:     source,
		ClassName: className,
	}, nil
}

// ClassifyImage 从字节数据分类图像，返回分类标签和最大概率值
func (l *TfClassificationLogic) ClassifyImage(imageBytes []byte) (string, float32, error) {

	// 解码字节数据为图像
	img, err := gocv.IMDecode(imageBytes, gocv.IMReadColor)
	if err != nil || img.Empty() {
		return "", 0, fmt.Errorf("failed to decode image: %v", err)
	}
	defer func(img *gocv.Mat) {
		_ = img.Close()
	}(&img)

	// 将图像 Mat 转换为 224x224 blob，以便分类器分析
	blob := gocv.BlobFromImage(img, 1.0, image.Pt(224, 224), gocv.NewScalar(0, 0, 0, 0), true, false)

	// 将 blob 输入分类器
	l.svcCtx.TfNet.SetInput(blob, "input")

	// 运行网络的正向传递
	prob := l.svcCtx.TfNet.Forward("softmax2")

	// 将结果重塑为 1x1000 矩阵
	probMat := prob.Reshape(1, 1)

	// 确定最可能的分类
	_, maxVal, _, maxLoc := gocv.MinMaxLoc(probMat)

	// 获取分类描述
	desc := ""
	if maxLoc.X < 1000 {
		desc = l.svcCtx.TfDesc[maxLoc.X]
	}

	// 清理资源
	_ = blob.Close()
	_ = prob.Close()
	_ = probMat.Close()

	return desc, maxVal, nil
}
