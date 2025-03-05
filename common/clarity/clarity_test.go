package clarity

import (
	"bytes"
	"image"
	"os"
	"testing"
)

func TestClarity(t *testing.T) {

	//detector := NewDetector(
	//	WithConcurrency(8), WithBaseThreshold(90), WithEdgeBoost(1.2), WithSampleScale(1))
	//imgData, _ := os.ReadFile("4.png")
	//blurred, confidence, err := detector.Detect(imgData)
	//if err != nil {
	//	t.Error(err)
	//}
	//t.Log(blurred, confidence)
	imgData, _ := os.ReadFile("2.png")
	img, _, err := image.Decode(bytes.NewReader(imgData))
	clarity, err := Clarity(img)
	if err != nil {
		t.Error(err)
	}
	t.Log(clarity)
}
