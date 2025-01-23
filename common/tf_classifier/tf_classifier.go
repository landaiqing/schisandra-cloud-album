package tf_classifier

import (
	"bufio"
	"fmt"
	"gocv.io/x/gocv"
	"os"
	"path/filepath"
)

// NewTFClassifier 创建一个新的TFClassifier实例
func NewTFClassifier() (*gocv.Net, []string) {
	var err error
	dir, err := os.Getwd()
	// 加载模型
	model := filepath.Join(dir, "/resources/models/tf_classifier/inception5h/tensorflow_inception_graph.pb")
	description := filepath.Join(dir, "/resources/models/tf_classifier/inception5h/imagenet_comp_graph_label_strings.txt")

	net := gocv.ReadNet(model, "")
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
	descriptions, err := ReadDescriptions(description)
	if err != nil {
		panic(fmt.Errorf("error reading descriptions: %v", err))
	}
	return &net, descriptions
}

// ReadDescriptions 从文件中读取描述并返回其行的切片
func ReadDescriptions(path string) ([]string, error) {
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
