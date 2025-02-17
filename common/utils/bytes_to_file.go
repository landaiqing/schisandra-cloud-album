package utils

import (
	"fmt"
	"io"
	"mime/multipart"
)

// ByteReader 实现了 multipart.File 接口
type ByteReader struct {
	data  []byte
	index int
}

func (r *ByteReader) Read(p []byte) (n int, err error) {
	if r.index >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.index:])
	r.index += n
	return n, nil
}

func (r *ByteReader) Seek(offset int64, whence int) (int64, error) {
	return 0, fmt.Errorf("Seek not implemented")
}

func (r *ByteReader) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, fmt.Errorf("ReadAt not implemented")
}

// 实现 Close 方法，符合 multipart.File 接口
func (r *ByteReader) Close() error {
	// 这里没有实际需要清理的资源，但必须实现 Close 方法
	return nil
}

// ToMultipartFile 将 []byte 转换为 multipart.File
func ToMultipartFile(data []byte) multipart.File {
	return &ByteReader{data: data}
}
