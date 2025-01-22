package utils

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg" // 引入 jpeg 解码器
	_ "image/png"  // 引入 png 解码器
	"io"
	"regexp"
	"strings"
	"sync"
)

var wg sync.WaitGroup

// GetMimeType 获取 MIME 类型
func GetMimeType(data []byte) string {
	if len(data) < 4 {
		return "application/octet-stream" // 默认类型
	}

	// 判断 JPEG
	if data[0] == 0xFF && data[1] == 0xD8 {
		return "image/jpeg"
	}

	// 判断 PNG
	if len(data) >= 8 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 &&
		data[4] == 0x0D && data[5] == 0x0A && data[6] == 0x1A && data[7] == 0x0A {
		return "image/png"
	}

	// 判断 GIF
	if len(data) >= 6 && data[0] == 'G' && data[1] == 'I' && data[2] == 'F' {
		return "image/gif"
	}
	// 判断 WEBP
	if len(data) >= 12 && data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 &&
		data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50 {
		return "image/webp"
	}
	// 判断svg
	if len(data) >= 4 && data[0] == '<' && data[1] == '?' && data[2] == 'x' && data[3] == 'm' {
		return "image/svg+xml"
	}
	// 判断JPG
	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "image/jpeg"
	}

	return "application/octet-stream" // 默认类型
}

// ProcessImages 处理图片，将 base64 字符串转换为字节数组
func ProcessImages(images []string) ([][]byte, error) {
	var imagesData [][]byte
	dataChan := make(chan []byte, len(images))
	re := regexp.MustCompile(`^data:image/\w+;base64,`)

	for _, img := range images {
		wg.Add(1)
		go func(img string) {
			defer wg.Done()

			imgWithoutPrefix := re.ReplaceAllString(img, "")
			data, err := Base64ToBytes(imgWithoutPrefix)
			if err != nil {
				return // 出错时直接返回
			}
			dataChan <- data
		}(img)
	}

	wg.Wait()
	close(dataChan)

	for data := range dataChan {
		imagesData = append(imagesData, data)
	}

	return imagesData, nil
}

// Base64ToBytes 将base64字符串转换为字节数组
func Base64ToBytes(base64Str string) ([]byte, error) {
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(base64Str))
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.New("failed to decode base64 string")
	}
	return data, nil
}

// Base64ToImage 将 Base64 字符串转换为 image.Image 格式
func Base64ToImage(base64Str string) (image.Image, error) {
	// 使用正则表达式去除前缀
	re := regexp.MustCompile(`^data:image/([a-zA-Z]*);base64,`)
	// 去除前缀部分
	base64Str = re.ReplaceAllString(base64Str, "")

	// 解码 Base64 字符串
	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 string: %v", err)
	}

	// 使用 image.Decode 解码字节数据
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	return img, nil
}
