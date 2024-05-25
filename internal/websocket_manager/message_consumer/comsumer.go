package message_consumer

import (
	"context"
	"encoding/json"
	"graduation-thesis/internal/websocket_manager/model"
	"graduation-thesis/internal/websocket_manager/service"
	"graduation-thesis/pkg/logger"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type MessageConsumer struct {
	consumer                *kafka.Consumer
	userService             *service.UserService
	websocketManagerService *service.WebsocketManagerService
	maxRetries              int
	retryInterval           time.Duration
	logger                  logger.Logger
}

func NewMessageConsumer(
	consumer *kafka.Consumer,
	userService *service.UserService,
	websocketManagerService *service.WebsocketManagerService,
	maxRetries int,
	retryInterval time.Duration,
	logger logger.Logger) *MessageConsumer {
	return &MessageConsumer{
		consumer:                consumer,
		userService:             userService,
		websocketManagerService: websocketManagerService,
		maxRetries:              maxRetries,
		retryInterval:           retryInterval,
		logger:                  logger,
	}
}

func rebalanceCallBack(*kafka.Consumer, kafka.Event) error {
	return nil
}

func (m *MessageConsumer) ListenAndServe(topics []string) error {
	m.logger.Info("[Consumer] Starting Message Consumer")
	err := m.consumer.SubscribeTopics(topics, rebalanceCallBack)
	if err != nil {
		return err
	}

	for {
		event := m.consumer.Poll(100)
		switch e := event.(type) {
		case *kafka.Message:
			m.logger.Infof("[Consumer] Message at %d[%d]: %v\n", e.TopicPartition.Partition, e.TopicPartition.Offset, e)
			m.processMessage(e)
			m.logger.Infof("[Consumer] Done in processing message at %d[%d]\n", e.TopicPartition.Partition, e.TopicPartition.Offset)
		case kafka.PartitionEOF:
			m.logger.Debugf("[Consumer] Reached: %v\n", e)
		case kafka.Error:
			m.logger.Errorf("[Consumer] Error: %v\n", e)
		default:
			m.logger.Debugf("[Consumer] Ignored: %v\n", e)
		}
	}
}

func (m *MessageConsumer) processMessage(message *kafka.Message) {
	ctx := context.Background()
	var kafkaMessage model.KafkaMessage
	if err := json.Unmarshal(message.Value, &kafkaMessage); err != nil {
		m.logger.Errorf("[Consumer] Cannot unmarshal message at %d[%d]: %v\n", message.TopicPartition.Partition, message.TopicPartition.Offset, err)
		return
	}

	if kafkaMessage.Action == "add" {
		for i := 1; i <= m.maxRetries; i++ {
			addNewUserRequest := model.AddNewUserRequest{
				UserID:      kafkaMessage.UserID,
				WebsocketID: kafkaMessage.WebsocketHandlerID,
			}
			_, errorResponse := m.websocketManagerService.AddNewUser(ctx, &addNewUserRequest)
			if errorResponse == nil {
				break
			}
			m.logger.Errorf("[Consumer] Failed to add user for %dth time: %v", i, errorResponse.ErrorMessage)
			time.Sleep(m.retryInterval)
		}
	}

	if kafkaMessage.Action == "remove" {
		for i := 1; i <= m.maxRetries; i++ {
			addNewUserRequest := model.AddNewUserRequest{
				UserID:      kafkaMessage.UserID,
				WebsocketID: kafkaMessage.WebsocketHandlerID,
			}
			_, errorResponse := m.websocketManagerService.RemoveUser(ctx, &addNewUserRequest)
			if errorResponse == nil {
				break
			}
			m.logger.Errorf("[Consumer] Failed to remove user for %dth time: %v", i, errorResponse.ErrorMessage)
			time.Sleep(m.retryInterval)
		}
	}
}
