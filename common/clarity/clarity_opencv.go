package clarity

import (
	"gocv.io/x/gocv"
	"image"
)

// 清晰度检测
func Clarity(img image.Image) (bool, error) {
	mat, err := gocv.ImageToMatRGB(img)
	if err != nil || mat.Empty() {
		if mat.Empty() == false {
			mat.Close()
		}
		return false, err
	}
	matClone := mat.Clone()
	// 如果图片是多通道 就进去转换
	if mat.Channels() != 1 {
		// 将图像转换为灰度显示
		gocv.CvtColor(mat, &matClone, gocv.ColorRGBToGray)
	}
	mat.Close()

	destCanny := gocv.NewMat()

	destCannyC := gocv.NewMat()

	destCannyD := gocv.NewMat()
	// 边缘检测
	gocv.Canny(matClone, &destCanny, 200, 200)
	// 求矩阵的均值与标准差
	gocv.MeanStdDev(destCanny, &destCannyC, &destCannyD)
	destCanny.Close()
	destCannyC.Close()
	if destCannyD.GetDoubleAt(0, 0) == 0 {
		destCannyD.Close()
		matClone.Close()
		return false, nil
	}
	destCannyD.Close()

	destC := gocv.NewMat()
	destD := gocv.NewMat()
	destA := gocv.NewMat()
	// Laplace算子
	gocv.Laplacian(matClone, &destA, gocv.MatTypeCV64F, 3, 1, 0, gocv.BorderDefault)
	gocv.MeanStdDev(destA, &destC, &destD)
	destC.Close()
	destA.Close()
	destMean := gocv.NewMat()
	gocv.Laplacian(matClone, &destMean, gocv.MatTypeCV16U, 3, 1, 0, gocv.BorderDefault)
	mean := destMean.Mean()
	destMean.Close()
	matClone.Close()
	if mean.Val1 > 5 || destD.GetDoubleAt(0, 0) > 20 {
		destD.Close()
		return true, nil
	}
	destD.Close()
	return false, nil
}
