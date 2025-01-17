package utils

import (
	"crypto"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
)

// SupportedHashFuncs 定义支持的哈希函数类型
var SupportedHashFuncs = map[string]func() hash.Hash{
	"md5":    crypto.MD5.New,
	"sha1":   crypto.SHA1.New,
	"sha256": crypto.SHA256.New,
	"sha512": crypto.SHA512.New,
}

// CalculateFileHash 根据指定的哈希算法计算文件的哈希值
func CalculateFileHash(filePath string, algorithm string) (string, error) {
	// 获取对应的哈希函数
	hashFunc, exists := SupportedHashFuncs[algorithm]
	if !exists {
		return "", fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 创建哈希对象
	h := hashFunc()

	// 计算哈希值
	if _, err := io.Copy(h, file); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}

	// 返回哈希值的十六进制字符串
	return hex.EncodeToString(h.Sum(nil)), nil
}

// CalculateStreamHash 计算输入流的哈希值
func CalculateStreamHash(reader io.Reader, algorithm string) (string, error) {
	// 获取对应的哈希函数
	hashFunc, exists := SupportedHashFuncs[algorithm]
	if !exists {
		return "", fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}

	// 创建哈希对象
	h := hashFunc()

	// 从输入流计算哈希值
	if _, err := io.Copy(h, reader); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}

	// 返回哈希值的十六进制字符串
	return hex.EncodeToString(h.Sum(nil)), nil
}
