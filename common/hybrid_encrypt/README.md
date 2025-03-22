# 混合加密图片方案

本模块提供了一套完整的混合加密解决方案，用于图片数据的安全传输。它使用RSA和AES-GCM混合加密方式，确保数据传输的安全性和效率。

## 工作原理

1. **混合加密**：使用RSA加密AES密钥，使用AES-GCM加密实际数据
   - 生成随机AES-256密钥
   - 使用RSA公钥加密AES密钥
   - 使用AES-GCM加密图片数据
   - 组合加密后的AES密钥、nonce和加密数据
   - 将结果进行Base64编码，便于网络传输

2. **混合解密**：使用RSA私钥解密AES密钥，使用AES-GCM解密数据
   - Base64解码密文
   - 分离加密的AES密钥、nonce和加密数据
   - 使用RSA私钥解密AES密钥
   - 使用解密后的AES密钥和nonce解密数据

## 后端使用方法 (Go)

```go
// 生成RSA密钥对
privateKey, _ := hybrid_encrypt.GenerateRSAKey(2048)

// 导出公钥和私钥为PEM格式
pubPEM, _ := hybrid_encrypt.ExportPublicKeyPEM(&privateKey.PublicKey)
privPEM := hybrid_encrypt.ExportPrivateKeyPEM(privateKey)

// 将PEM格式的密钥转换为Base64编码，便于在网络上传输
base64PubKey := base64.StdEncoding.EncodeToString(pubPEM)
base64PrivKey := base64.StdEncoding.EncodeToString(privPEM)

// 读取图片文件
imageData, _ := ioutil.ReadFile("path/to/image.jpg")

// 使用公钥加密图片
encryptedBase64, _ := hybrid_encrypt.EncryptImageWithBase64Key(base64PubKey, imageData)

// 将加密后的Base64字符串发送到前端
// ...

// 如果需要在后端解密
decryptedData, _ := hybrid_encrypt.DecryptImageWithBase64Key(base64PrivKey, encryptedBase64)
```

## 前端使用方法 (JavaScript)

```javascript
import { decryptImage } from './hybrid_encrypt_js.js';

// 从后端接收加密的图片数据和私钥
const encryptedImageBase64 = '...'; // 从后端接收的Base64编码的加密图片数据
const privateKeyPEM = '...'; // 从安全渠道获取的PEM格式RSA私钥

// 解密图片
decryptImage(encryptedImageBase64, privateKeyPEM)
  .then(decryptedImageBase64 => {
    // 使用解密后的图片数据
    const imgElement = document.createElement('img');
    imgElement.src = decryptedImageBase64; // 直接设置为data:image/jpeg;base64,...格式
    document.body.appendChild(imgElement);
  })
  .catch(error => {
    console.error('图片解密失败:', error);
  });
```

## 安全注意事项

1. 私钥必须妥善保管，不应在不安全的渠道传输
2. 对于大型图片，考虑分块加密和解密
3. 在生产环境中，应使用至少2048位的RSA密钥
4. 前端解密应在安全的环境中进行，避免私钥泄露

## 性能考虑

混合加密方案在处理大型数据（如高分辨率图片）时具有明显的性能优势：

- RSA仅用于加密小型AES密钥（32字节）
- 大型图片数据使用高效的AES-GCM加密
- 加密后的数据大小增加有限，主要是RSA加密的AES密钥部分