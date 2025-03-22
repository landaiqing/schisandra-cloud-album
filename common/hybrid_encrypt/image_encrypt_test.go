package hybrid_encrypt

import (
	"fmt"
	"testing"
)

// 测试生成RSA密钥对、导出公钥和私钥为PEM格式、加密解密图片数据
// 注意：测试代码仅用于演示，实际应用中请使用更安全的加密算法和密钥长度
// 请不要在生产环境中使用此测试代码
func TestHybridEncrypt(t *testing.T) {
	// 生成RSA密钥对
	privateKey, err := GenerateRSAKey(2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	// 导出公钥和私钥为PEM格式
	pubPEM, err := ExportPublicKeyPEM(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to export public key: %v", err)
	}

	privPEM := ExportPrivateKeyPEM(privateKey)
	// 打印公钥和私钥的PEM格式
	fmt.Println(string(pubPEM))
	fmt.Println(string(privPEM))
}
