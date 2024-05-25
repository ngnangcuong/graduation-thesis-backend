package storage

import "github.com/confluentinc/confluent-kafka-go/v2/kafka"

func NewKafkaConsumer(bootstrapServers, groupID string) *kafka.Consumer {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"group.id":          groupID,
		"auto.offset.reset": "smallest"})
	if err != nil {
		panic(err)
	}

	return consumer
}
