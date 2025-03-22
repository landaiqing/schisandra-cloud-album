package hybrid_encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
)

// HybridEncrypt 使用RSA公钥加密AES密钥，并用AES-GCM加密数据
func HybridEncrypt(publicKey *rsa.PublicKey, plaintext []byte) ([]byte, error) {
	// 生成随机AES-256密钥
	aesKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		return nil, err
	}

	// 创建AES-GCM实例
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 生成随机Nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// 加密数据
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// 加密AES密钥
	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, aesKey, nil)
	if err != nil {
		return nil, err
	}

	// 组合最终密文：加密的AES密钥 + nonce + AES加密的数据
	result := make([]byte, len(encryptedKey)+len(nonce)+len(ciphertext))
	copy(result[:len(encryptedKey)], encryptedKey)
	copy(result[len(encryptedKey):len(encryptedKey)+len(nonce)], nonce)
	copy(result[len(encryptedKey)+len(nonce):], ciphertext)

	return result, nil
}

// HybridDecrypt 使用RSA私钥解密AES密钥，并用AES-GCM解密数据
func HybridDecrypt(privateKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	keySize := privateKey.PublicKey.Size()
	if len(ciphertext) < keySize+12 {
		return nil, errors.New("ciphertext too short")
	}

	// 分解密文各部分
	encryptedKey := ciphertext[:keySize]
	nonce := ciphertext[keySize : keySize+12]
	data := ciphertext[keySize+12:]

	// 解密AES密钥
	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedKey, nil)
	if err != nil {
		return nil, err
	}

	// 创建AES-GCM实例
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 解密数据
	plaintext, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// GenerateRSAKey 生成RSA密钥对
func GenerateRSAKey(bits int) (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, bits)
}

// ExportPublicKeyPEM 导出公钥为PEM格式
func ExportPublicKeyPEM(pub *rsa.PublicKey) ([]byte, error) {
	pubASN1, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, err
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	})
	return pubPEM, nil
}

// ExportPrivateKeyPEM 导出私钥为PEM格式
func ExportPrivateKeyPEM(priv *rsa.PrivateKey) []byte {
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})
	return privPEM
}

// ImportPublicKeyPEM 从PEM导入公钥
func ImportPublicKeyPEM(pubPEM []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pubPEM)
	if block == nil {
		return nil, errors.New("failed to parse PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not RSA public key")
	}
	return rsaPub, nil
}

// ImportPrivateKeyPEM 从PEM导入私钥
func ImportPrivateKeyPEM(privPEM []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privPEM)
	if block == nil {
		return nil, errors.New("failed to parse PEM block")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func main() {
	// 生成RSA密钥对
	privateKey, err := GenerateRSAKey(2048)
	if err != nil {
		panic(err)
	}

	// 导出导入测试
	pubPEM, _ := ExportPublicKeyPEM(&privateKey.PublicKey)
	publicKey, _ := ImportPublicKeyPEM(pubPEM)

	privPEM := ExportPrivateKeyPEM(privateKey)
	privateKey, _ = ImportPrivateKeyPEM(privPEM)

	// 加密测试
	message := []byte("Secret image data")
	fmt.Printf("Original: %s\n", message)

	ciphertext, err := HybridEncrypt(publicKey, message)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Ciphertext length: %d bytes\n", len(ciphertext))

	// 解密测试
	plaintext, err := HybridDecrypt(privateKey, ciphertext)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Decrypted: %s\n", plaintext)
}
