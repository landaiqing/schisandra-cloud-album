package encrypt

import (
	"fmt"
	"log"
	"testing"
)

func TestAES(t *testing.T) {
	key := "thisisasecretkey" // 16 字节密钥
	plainText := "Hello, AES-GCM encryption!"

	// 加密
	encrypted, err := Encrypt(plainText, key)
	if err != nil {
		log.Fatalf("加密失败: %v", err)
	}
	fmt.Printf("加密结果: %s\n", encrypted)

	// 解密
	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		log.Fatalf("解密失败: %v", err)
	}
	fmt.Printf("解密结果: %s\n", decrypted)
}
