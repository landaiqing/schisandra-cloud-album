package clarity

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"testing"
)

func TestClarity(t *testing.T) {

	imgData, err := os.Open("2.jpg")
	if err != nil {
		t.Error(err)
	}
	img, _, err := image.Decode(imgData)
	if err != nil {
		t.Error(err)
	}
	detector := NewConcurrentDetector(WithMeanThreshold(13.0), // 提高均值阈值
		WithLaplaceStdThreshold(25.0), // 提高标准差阈值
		WithMaxWorkers(8),             // 设置并发数
	)
	check, err := detector.ClarityCheck(img)
	if err != nil {
		t.Error(err)
	}
	t.Log(check)
}
