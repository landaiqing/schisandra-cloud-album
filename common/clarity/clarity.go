package clarity

import (
	"context"
	"fmt"
	"image"
	"runtime"
	"sync"

	"gocv.io/x/gocv"
	"golang.org/x/sync/semaphore"
)

// 模糊检测器
type ConcurrentDetector struct {
	maxWorkers          int64   // 最大并发数
	meanThreshold       float64 // 均值阈值
	laplaceStdThreshold float64 // Laplace标准差阈值
	sem                 *semaphore.Weighted
}

type Option func(*ConcurrentDetector)

// 默认参数
func NewConcurrentDetector(opts ...Option) *ConcurrentDetector {
	d := &ConcurrentDetector{
		maxWorkers:          int64(runtime.NumCPU() * 2),
		meanThreshold:       5.0,  // 原始值5
		laplaceStdThreshold: 20.0, // 原始值20
	}

	for _, opt := range opts {
		opt(d)
	}

	d.sem = semaphore.NewWeighted(d.maxWorkers)
	return d
}

// 配置选项 -------------------------------------------------
func WithMeanThreshold(t float64) Option {
	return func(d *ConcurrentDetector) {
		d.meanThreshold = t
	}
}

func WithLaplaceStdThreshold(t float64) Option {
	return func(d *ConcurrentDetector) {
		d.laplaceStdThreshold = t
	}
}

func WithMaxWorkers(n int) Option {
	return func(d *ConcurrentDetector) {
		d.maxWorkers = int64(n)
	}
}

func (d *ConcurrentDetector) ClarityCheck(img image.Image) (bool, error) {
	if img == nil {
		return false, fmt.Errorf("nil image input")
	}

	mat, err := gocv.ImageToMatRGB(img)
	if err != nil || mat.Empty() {
		if mat.Empty() == false {
			mat.Close()
		}
		return false, err
	}
	matClone := mat.Clone()
	if mat.Channels() != 1 {
		gocv.CvtColor(mat, &matClone, gocv.ColorRGBToGray)
	}
	mat.Close()

	// Canny检测部分
	destCanny := gocv.NewMat()
	defer destCanny.Close()
	gocv.Canny(matClone, &destCanny, 200, 200)

	destCannyC := gocv.NewMat()
	defer destCannyC.Close()
	destCannyD := gocv.NewMat()
	defer destCannyD.Close()
	gocv.MeanStdDev(destCanny, &destCannyC, &destCannyD)
	if destCannyD.GetDoubleAt(0, 0) == 0 {
		matClone.Close()
		return false, nil
	}

	// Laplace检测部分
	destA := gocv.NewMat()
	defer destA.Close()
	gocv.Laplacian(matClone, &destA, gocv.MatTypeCV64F, 3, 1, 0, gocv.BorderDefault)

	destC := gocv.NewMat()
	defer destC.Close()
	destD := gocv.NewMat()
	defer destD.Close()
	gocv.MeanStdDev(destA, &destC, &destD)

	destMean := gocv.NewMat()
	defer destMean.Close()
	gocv.Laplacian(matClone, &destMean, gocv.MatTypeCV16U, 3, 1, 0, gocv.BorderDefault)
	mean := destMean.Mean()
	matClone.Close()

	// 使用可配置阈值（mean.Val1 >5 || destD.GetDoubleAt>20）
	result := mean.Val1 > d.meanThreshold && destD.GetDoubleAt(0, 0) > d.laplaceStdThreshold
	return result, nil
}

type Result struct {
	Blurred bool
	Err     error
}

func (d *ConcurrentDetector) BatchDetect(ctx context.Context, images <-chan image.Image) <-chan Result {
	results := make(chan Result)
	var wg sync.WaitGroup

	go func() {
		defer close(results)
		for img := range images {
			if err := d.sem.Acquire(ctx, 1); err != nil {
				break
			}
			wg.Add(1)

			go func(img image.Image) {
				defer wg.Done()
				defer d.sem.Release(1)

				blurred, err := d.ClarityCheck(img)
				select {
				case results <- Result{Blurred: blurred, Err: err}:
				case <-ctx.Done():
				}
			}(img)
		}
		wg.Wait()
	}()

	return results
}

/*
func main() {
	// 初始化检测器（调整阈值参数）
	detector := NewConcurrentDetector(
		WithMeanThreshold(8.0),        // 提高均值阈值
		WithLaplaceStdThreshold(25.0),  // 提高标准差阈值
		WithMaxWorkers(8),              // 设置并发数
	)

	// 准备测试图片
	img := loadImage("test.jpg")

	// 单张检测
	blurred, _ := detector.clarityCheck(img)
	fmt.Println("Blurred:", blurred)

	// 批量检测
	ctx := context.Background()
	imgChan := make(chan image.Image, 10)
	go func() {
		for i := 0; i < 10; i++ {
			imgChan <- loadImage(fmt.Sprintf("image%d.jpg", i))
		}
		close(imgChan)
	}()

	results := detector.BatchDetect(ctx, imgChan)
	for res := range results {
		if res.Err != nil {
			fmt.Println("Error:", res.Err)
			continue
		}
		fmt.Println("Result:", res.Blurred)
	}
}
*/
