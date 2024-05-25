package storage

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

var kafkaProducer *kafka.Producer

func NewKafkaProducer(bootstrapServers string, messageMaxBytes int) *kafka.Producer {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"message.max.bytes": messageMaxBytes,
	})
	if err != nil {
		panic(err)
	}

	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	return producer
}

func GetKafkaProducer(bootstrapServers string, messageMaxBytes int) *kafka.Producer {
	if kafkaProducer == nil {
		kafkaProducer = NewKafkaProducer(bootstrapServers, messageMaxBytes)
	}
	return kafkaProducer
}
