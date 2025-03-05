package clarity

import (
	"bytes"
	"context"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"runtime"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// Detector 图片模糊检测器
type Detector struct {
	baseThreshold     float64    // 基准阈值
	sampleScale       int        // 采样比例基数
	edgeBoost         float64    // 边缘增强系数
	noiseFloor        float64    // 噪声基底
	channelWeights    [3]float64 // RGB通道权重
	adaptiveSampling  bool       // 启用自适应采样
	regionWeights     []float64  // 区域权重矩阵
	concurrencyLimit  int64      // 最大并发数
	weightedSemaphore *semaphore.Weighted
	pool              sync.Pool // 内存池
}

type Option func(*Detector)

// NewDetector 创建检测器实例
func NewDetector(opts ...Option) *Detector {
	d := &Detector{
		baseThreshold:    85.0,
		sampleScale:      2,
		edgeBoost:        1.0,
		noiseFloor:       5.0,
		channelWeights:   [3]float64{0.299, 0.587, 0.114},
		adaptiveSampling: true,
		concurrencyLimit: int64(runtime.NumCPU() * 2),
	}

	d.pool.New = func() interface{} {
		return &scanContext{
			sum:   0,
			sumSq: 0,
		}
	}

	d.weightedSemaphore = semaphore.NewWeighted(d.concurrencyLimit)

	for _, opt := range opts {
		opt(d)
	}
	return d
}

// 配置选项 ---------------------------------------------------

func WithBaseThreshold(t float64) Option {
	return func(d *Detector) {
		d.baseThreshold = t
	}
}

func WithSampleScale(n int) Option {
	return func(d *Detector) {
		d.sampleScale = 1 << uint(maxInt(0, n))
	}
}

func WithEdgeBoost(factor float64) Option {
	return func(d *Detector) {
		d.edgeBoost = clamp(factor, 0.5, 2.0)
	}
}

func WithNoiseFloor(floor float64) Option {
	return func(d *Detector) {
		d.noiseFloor = math.Max(0, floor)
	}
}

func WithConcurrency(n int) Option {
	return func(d *Detector) {
		d.concurrencyLimit = int64(maxInt(1, n))
		d.weightedSemaphore = semaphore.NewWeighted(d.concurrencyLimit)
	}
}

// Detect 执行模糊检测
func (d *Detector) Detect(imgData []byte) (isBlurred bool, confidence float64, err error) {
	img, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return true, 0.0, err
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	if width < 32 || height < 32 {
		return true, 0.0, nil
	}

	ctx := d.pool.Get().(*scanContext)
	defer d.pool.Put(ctx)
	ctx.reset()

	step := d.calculateStep(width, height)

	g, groupCtx := errgroup.WithContext(context.Background())
	processingCtx, cancel := context.WithCancel(groupCtx)
	defer cancel()

	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		for x := bounds.Min.X; x < bounds.Max.X; x += step {
			x, y := x, y // 捕获循环变量

			if err := d.weightedSemaphore.Acquire(processingCtx, 1); err != nil {
				break
			}

			g.Go(func() error {
				defer d.weightedSemaphore.Release(1)

				select {
				case <-processingCtx.Done():
					return nil
				default:
				}

				if x <= 0 || y <= 0 || x >= bounds.Max.X-1 || y >= bounds.Max.Y-1 {
					return nil
				}

				gray := d.calculateGray(img, x, y)
				val := d.calculateLaplacian(img, x, y, gray)
				weight := d.getRegionWeight(x, y, bounds)

				ctx.mu.Lock()
				ctx.sum += val * weight
				ctx.sumSq += (val * weight) * (val * weight)
				ctx.mu.Unlock()

				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		return true, 0.0, err
	}

	n := float64(((width / step) * (height / step)) - 4)
	if n <= 0 {
		return true, 0.0, nil
	}

	mean := ctx.sum / n
	variance := (ctx.sumSq/n - mean*mean) * 1e6

	dynamicThreshold := d.calculateDynamicThreshold(width, height)
	confidence = math.Max(0, math.Min(1, (variance-d.noiseFloor)/(dynamicThreshold-d.noiseFloor)))

	return variance < dynamicThreshold, confidence, nil
}

// 私有方法 ---------------------------------------------------

func (d *Detector) calculateStep(width, height int) int {
	if !d.adaptiveSampling {
		return d.sampleScale
	}

	area := width * height
	switch {
	case area > 4000*3000:
		return d.sampleScale * 4
	case area > 2000*1500:
		return d.sampleScale * 2
	default:
		return d.sampleScale
	}
}

func (d *Detector) calculateGray(img image.Image, x, y int) float64 {
	r, g, b, _ := img.At(x, y).RGBA()
	return d.channelWeights[0]*float64(r>>8) +
		d.channelWeights[1]*float64(g>>8) +
		d.channelWeights[2]*float64(b>>8)
}

func (d *Detector) calculateLaplacian(img image.Image, x, y int, center float64) float64 {
	getGray := func(x, y int) float64 {
		r, g, b, _ := img.At(x, y).RGBA()
		return d.channelWeights[0]*float64(r>>8) +
			d.channelWeights[1]*float64(g>>8) +
			d.channelWeights[2]*float64(b>>8)
	}

	return math.Abs(4*center-
		getGray(x-1, y)-
		getGray(x+1, y)-
		getGray(x, y-1)-
		getGray(x, y+1)) * d.edgeBoost
}

func (d *Detector) calculateDynamicThreshold(width, height int) float64 {
	areaRatio := float64(width*height) / 250000.0
	return d.baseThreshold*math.Pow(areaRatio, 0.65) + d.noiseFloor
}

func (d *Detector) getRegionWeight(x, y int, bounds image.Rectangle) float64 {
	if len(d.regionWeights) == 0 {
		return 1.0
	}

	size := int(math.Sqrt(float64(len(d.regionWeights))))
	if size == 0 {
		return 1.0
	}

	nx := float64(x-bounds.Min.X) / float64(bounds.Dx())
	ny := float64(y-bounds.Min.Y) / float64(bounds.Dy())

	ix := int(nx * float64(size))
	iy := int(ny * float64(size))
	idx := iy*size + ix

	if idx >= 0 && idx < len(d.regionWeights) {
		return d.regionWeights[idx]
	}
	return 1.0
}

// 辅助函数 ---------------------------------------------------

type scanContext struct {
	sum   float64
	sumSq float64
	mu    sync.Mutex
}

func (c *scanContext) reset() {
	c.sum = 0
	c.sumSq = 0
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clamp(value, min, max float64) float64 {
	return math.Max(min, math.Min(max, value))
}
