package utils

import (
	"github.com/corona10/goimagehash"
	"image"
)

// CalculatePerceptualHash 计算感知哈希
func CalculatePerceptualHash(img image.Image) (string, error) {
	hash, err := goimagehash.PerceptionHash(img)
	if err != nil {
		return "", err
	}
	return hash.ToString(), nil
}

// CalculateHash 计算平均哈希
func CalculateHash(img image.Image) (uint64, error) {
	hash, err := goimagehash.AverageHash(img)
	if err != nil {
		return 0, err
	}
	return hash.GetHash(), nil
}
