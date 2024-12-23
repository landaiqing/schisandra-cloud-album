package idgenerator

import "github.com/yitter/idgenerator-go/idgen"

func NewIDGenerator(workerId uint16) {
	var options = idgen.NewIdGeneratorOptions(workerId)
	options.WorkerIdBitLength = 6 // 默认值6，限定 WorkerId 最大值为2^6-1，即默认最多支持64个节点。
	options.SeqBitLength = 6      // 默认值6，限制每毫秒生成的ID个数。若生成速度超过5万个/秒，建议加大 SeqBitLength 到 10。
	idgen.SetIdGenerator(options)
}
