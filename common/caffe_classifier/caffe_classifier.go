package caffe_classifier

import (
	"bufio"
	"fmt"
	"gocv.io/x/gocv"
	"os"
	"path/filepath"
)

// NewCaffeClassifier 创建一个新的Caffe分类器
func NewCaffeClassifier() (*gocv.Net, []string) {
	var err error
	dir, err := os.Getwd()
	// 加载模型
	model := filepath.Join(dir, "/resources/models/caffemodel/bvlc_googlenet.caffemodel")
	config := filepath.Join(dir, "/resources/models/caffemodel/bvlc_googlenet.prototxt")
	description := filepath.Join(dir, "/resources/models/caffemodel/classification_classes_ILSVRC2012.txt")

	net := gocv.ReadNet(model, config)
	if net.Empty() {
		panic(fmt.Errorf("error reading network model: %v", model))
	}
	// 设置后端和目标设备
	err = net.SetPreferableBackend(gocv.NetBackendDefault)
	if err != nil {
		panic(fmt.Errorf("error setting preferable backend: %v", err))
	}
	err = net.SetPreferableTarget(gocv.NetTargetCPU)
	if err != nil {
		panic(fmt.Errorf("error setting preferable target: %v", err))
	}
	// 加载描述文件
	descriptions, err := readDescriptions(description)
	if err != nil {
		panic(fmt.Errorf("error reading descriptions: %v", err))
	}
	return &net, descriptions
}

// readDescriptions reads the descriptions from a file
// and returns a slice of its lines.
func readDescriptions(path string) ([]string, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
