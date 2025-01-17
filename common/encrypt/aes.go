package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// Encrypt 使用 AES-GCM 模式加密
func Encrypt(plainText string, key string) (string, error) {
	// 转换 key 为字节数组
	keyBytes := []byte(key)

	// 创建 AES block
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	// 创建 GCM 实例
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 生成随机 nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 加密明文
	cipherText := aesGCM.Seal(nil, nonce, []byte(plainText), nil)

	// 将 nonce 和密文拼接后进行 Base64 编码
	result := append(nonce, cipherText...)
	return base64.StdEncoding.EncodeToString(result), nil
}

// Decrypt 使用 AES-GCM 模式解密
func Decrypt(cipherText string, key string) (string, error) {
	// 转换 key 为字节数组
	keyBytes := []byte(key)

	// Base64 解码密文
	data, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	// 创建 AES block
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	// 创建 GCM 实例
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 检查数据长度是否足够
	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("cipherText too short")
	}

	// 分离 nonce 和密文
	nonce, cipherTextBytes := data[:nonceSize], data[nonceSize:]

	// 解密密文
	plainText, err := aesGCM.Open(nil, nonce, cipherTextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}
