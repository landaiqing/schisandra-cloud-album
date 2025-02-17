package nsqx

import (
	"fmt"
	"github.com/nsqio/go-nsq"
	"time"
)

func NewNsqProducer(url string) *nsq.Producer {
	producer, err := nsq.NewProducer(url, nsq.NewConfig())
	if err != nil {
		panic(err)
	}
	producer.SetLoggerLevel(nsq.LogLevelError)
	return producer
}

func NewNSQConsumer(topic string) *nsq.Consumer {
	config := nsq.NewConfig()
	config.LookupdPollInterval = 15 * time.Second
	consumer, err := nsq.NewConsumer(topic, "channel", config)
	if err != nil {
		fmt.Printf("InitNSQ consumer error: %v\n", err)
		return nil
	}
	consumer.SetLoggerLevel(nsq.LogLevelError)
	return consumer
}
