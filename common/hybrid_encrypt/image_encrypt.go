package hybrid_encrypt

import (
	"crypto/rsa"
	"errors"
)

// EncryptImage 使用混合加密方式加密图片数据并返回Base64编码的密文
// 参数:
//   - publicKey: RSA公钥
//   - imageData: 原始图片数据
//
// 返回:
//   - string: Base64编码的加密图片数据
//   - error: 错误信息
func EncryptImage(publicKey *rsa.PublicKey, imageData []byte) ([]byte, error) {
	if len(imageData) == 0 {
		return nil, errors.New("empty image data")
	}

	// 使用混合加密方式加密图片数据
	ciphertext, err := HybridEncrypt(publicKey, imageData)
	if err != nil {
		return nil, err
	}

	return ciphertext, nil
}

// DecryptImage 解密Base64编码的加密图片数据
// 参数:
//   - privateKey: RSA私钥
//   - base64Ciphertext: Base64编码的加密图片数据
//
// 返回:
//   - []byte: 解密后的原始图片数据
//   - error: 错误信息
func DecryptImage(privateKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	// 使用混合解密方式解密图片数据
	imageData, err := HybridDecrypt(privateKey, ciphertext)
	if err != nil {
		return nil, err
	}

	return imageData, nil
}
