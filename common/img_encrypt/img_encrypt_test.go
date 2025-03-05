package main

import (
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	"testing"
)

func TestImgEncrypt(t *testing.T) {
	// 1. 读取并强制转换为RGBA
	inputFile, err := os.Open("E:\\Go_WorkSpace\\schisandra-album-cloud-microservices\\common\\img_encrypt\\input.png")
	if err != nil {
		log.Fatal("打开文件失败:", err)
	}
	defer inputFile.Close()

	srcImg, err := png.Decode(inputFile)
	if err != nil {
		log.Fatal("解码失败:", err)
	}

	bounds := srcImg.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, srcImg, bounds.Min, draw.Src)

	// 2. 安全加密（处理有效像素区）
	key := []byte{0x1F, 0x3A, 0x7B, 0x9C} // 示例密钥（推荐长度4/8/16）
	secureXor(rgba, key)

	// 3. 保存加密图像（禁用压缩）
	outputFile, err := os.Create("encrypted.png")
	if err != nil {
		log.Fatal("创建文件失败:", err)
	}
	defer outputFile.Close()

	encoder := png.Encoder{CompressionLevel: png.NoCompression}
	if err := encoder.Encode(outputFile, rgba); err != nil {
		log.Fatal("保存失败:", err)
	}
}

func TestImgDecrypt(t *testing.T) {
	// 1. 读取加密图像
	inputFile, err := os.Open("E:\\Go_WorkSpace\\schisandra-album-cloud-microservices\\common\\img_encrypt\\encrypted.png")
	if err != nil {
		log.Fatal("打开加密文件失败:", err)
	}
	defer inputFile.Close()

	encImg, err := png.Decode(inputFile)
	if err != nil {
		log.Fatal("解码失败:", err)
	}

	// 2. 转换为RGBA
	bounds := encImg.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, encImg, bounds.Min, draw.Src)

	// 3. 解密（使用相同密钥）
	key := []byte{0x1F, 0x3A, 0x7B, 0x9C} // 必须与加密一致
	secureXor(rgba, key)

	// 4. 保存解密结果
	outputFile, err := os.Create("decrypted.png")
	if err != nil {
		log.Fatal("创建解密文件失败:", err)
	}
	defer outputFile.Close()

	encoder := png.Encoder{CompressionLevel: png.NoCompression}
	if err := encoder.Encode(outputFile, rgba); err != nil {
		log.Fatal("保存失败:", err)
	}
}

// 通用加密/解密函数
// 安全加密函数
func secureXor(img *image.RGBA, key []byte) {
	keyLen := len(key)
	if keyLen == 0 {
		log.Fatal("密钥不能为空")
	}

	bounds := img.Bounds()
	data := img.Pix
	stride := img.Stride
	width := bounds.Dx() * 4 // 每行实际需要的字节数

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		rowStart := (y - bounds.Min.Y) * stride
		// 严格限定处理范围为有效像素区
		end := rowStart + width
		if end > len(data) {
			end = len(data)
		}

		for pos := rowStart; pos < end; pos++ {
			data[pos] ^= key[pos%keyLen]
		}
	}
}
