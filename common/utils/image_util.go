package utils

import (
	"encoding/base64"
	"errors"
	_ "image/jpeg" // 引入 jpeg 解码器
	_ "image/png"  // 引入 png 解码器
	"io"
	"regexp"
	"strings"
)

// Base64ToBytes 将base64字符串转换为字节数组
func Base64ToBytes(base64Str string) ([]byte, error) {
	re := regexp.MustCompile(`^data:image/\w+;base64,`)
	imgWithoutPrefix := re.ReplaceAllString(base64Str, "")
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgWithoutPrefix))
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.New("failed to decode base64 string")
	}
	return data, nil
}
