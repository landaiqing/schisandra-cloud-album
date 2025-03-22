/**
 * hybrid_encrypt_js.js
 * 前端混合解密实现，与Go后端的hybrid_encrypt.go配合使用
 */

/**
 * 使用RSA私钥和AES-GCM解密数据
 * @param {ArrayBuffer} ciphertext - 加密的数据
 * @param {CryptoKey} privateKey - RSA私钥
 * @returns {Promise<ArrayBuffer>} - 解密后的数据
 */
async function hybridDecrypt(ciphertext, privateKey) {
  // 将ArrayBuffer转换为Uint8Array以便处理
  const ciphertextArray = new Uint8Array(ciphertext);
  
  // 获取RSA密钥大小（字节）
  const keySize = getPrivateKeySize(privateKey);
  
  // 检查密文长度是否足够
  if (ciphertextArray.length < keySize + 12) {
    throw new Error('密文太短');
  }
  
  // 分解密文各部分
  const encryptedKey = ciphertextArray.slice(0, keySize);
  const nonce = ciphertextArray.slice(keySize, keySize + 12);
  const data = ciphertextArray.slice(keySize + 12);
  
  // 使用RSA私钥解密AES密钥
  const aesKeyBuffer = await window.crypto.subtle.decrypt(
    {
      name: 'RSA-OAEP',
      hash: { name: 'SHA-256' },
    },
    privateKey,
    encryptedKey
  );
  
  // 使用AES密钥和GCM模式解密数据
  const aesKey = await window.crypto.subtle.importKey(
    'raw',
    aesKeyBuffer,
    { name: 'AES-GCM', length: 256 },
    false,
    ['decrypt']
  );
  
  const plaintext = await window.crypto.subtle.decrypt(
    {
      name: 'AES-GCM',
      iv: nonce,
    },
    aesKey,
    data
  );
  
  return plaintext;
}

/**
 * 从PEM格式导入RSA私钥
 * @param {string} pemString - PEM格式的私钥字符串
 * @returns {Promise<CryptoKey>} - 导入的RSA私钥
 */
async function importPrivateKeyFromPEM(pemString) {
  // 移除PEM头尾和换行符
  const pemContents = pemString
    .replace('-----BEGIN RSA PRIVATE KEY-----', '')
    .replace('-----END RSA PRIVATE KEY-----', '')
    .replace(/\s+/g, '');
  
  // Base64解码
  const binaryDer = window.atob(pemContents);
  const derArray = new Uint8Array(binaryDer.length);
  
  for (let i = 0; i < binaryDer.length; i++) {
    derArray[i] = binaryDer.charCodeAt(i);
  }
  
  // 导入私钥
  const privateKey = await window.crypto.subtle.importKey(
    'pkcs8',
    derArray.buffer,
    {
      name: 'RSA-OAEP',
      hash: { name: 'SHA-256' },
    },
    false,
    ['decrypt']
  );
  
  return privateKey;
}

/**
 * 获取RSA私钥的大小（字节）
 * @param {CryptoKey} privateKey - RSA私钥
 * @returns {number} - 密钥大小（字节）
 */
function getPrivateKeySize(privateKey) {
  // 在实际应用中，可能需要从privateKey对象中提取密钥大小
  // 这里简化处理，假设使用2048位RSA密钥（256字节）
  return 256; // 2048位 / 8 = 256字节
}

/**
 * 解密Base64编码的混合加密图片数据
 * @param {string} base64Data - Base64编码的加密图片数据
 * @param {string} privateKeyPEM - PEM格式的RSA私钥
 * @returns {Promise<string>} - 解密后的图片Base64字符串
 */
async function decryptImage(base64Data, privateKeyPEM) {
  try {
    // 解码Base64
    const binaryString = window.atob(base64Data);
    const bytes = new Uint8Array(binaryString.length);
    for (let i = 0; i < binaryString.length; i++) {
      bytes[i] = binaryString.charCodeAt(i);
    }
    
    // 导入私钥
    const privateKey = await importPrivateKeyFromPEM(privateKeyPEM);
    
    // 解密数据
    const decryptedData = await hybridDecrypt(bytes.buffer, privateKey);
    
    // 将解密后的图片数据转换为Base64
    const decryptedArray = new Uint8Array(decryptedData);
    let binary = '';
    for (let i = 0; i < decryptedArray.length; i++) {
      binary += String.fromCharCode(decryptedArray[i]);
    }
    
    // 返回带有MIME类型的Base64图片数据
    // 注意：这里假设是JPEG图片，实际应用中可能需要根据图片类型动态设置
    return 'data:image/jpeg;base64,' + window.btoa(binary);
  } catch (error) {
    console.error('图片解密失败:', error);
    throw error;
  }
}

// 导出函数
export {
  hybridDecrypt,
  importPrivateKeyFromPEM,
  decryptImage
};