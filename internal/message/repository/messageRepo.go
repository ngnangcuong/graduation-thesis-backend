package repository

import (
	"context"
	"encoding/json"
	"errors"
	"graduation-thesis/internal/message/model"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gocql/gocql"
)

type MessageRepo struct {
	session    *gocql.Session
	producer   *kafka.Producer
	kafkaTopic string
}

func NewMessageRepo(session *gocql.Session, producer *kafka.Producer, kafkaTopic string) *MessageRepo {
	return &MessageRepo{
		session:    session,
		producer:   producer,
		kafkaTopic: kafkaTopic,
	}
}

func (m *MessageRepo) Get(ctx context.Context, from, to string, limit int) ([]*model.Message, error) {
	query := `SELECT id, from, to, content, delivered_at, last_updated, received_at, read_at, deleted, status, type, link_to FROM messages
				WHERE from = ?, to = ? limit ?`
	scanner := m.session.Query(query, from, to, limit).WithContext(ctx).Iter().Scanner()

	var messages []*model.Message
	for scanner.Next() {
		var message model.Message
		err := scanner.Scan(&message)
		if err != nil {
			return nil, err
		}

		messages = append(messages, &message)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (m *MessageRepo) GetUnRead(ctx context.Context, from, to string, limit int) ([]*model.Message, error) {
	query := `SELECT id, from, to, content, delivered_at, last_updated, received_at, read_at, deleted, status, type, link_to FROM messages
				WHERE from = ? AND to = ? AND status = "deliverd" limit ?`
	scanner := m.session.Query(query, from, to, limit).WithContext(ctx).Iter().Scanner()

	var messages []*model.Message
	for scanner.Next() {
		var message model.Message
		err := scanner.Scan(&message)
		if err != nil {
			return nil, err
		}

		messages = append(messages, &message)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (m *MessageRepo) GetConversation(ctx context.Context, users []string, timeFrom, timeTo time.Time, limit int) ([]*model.Message, error) {
	query := `SELECT id, from, to, content, delivered_at, last_updated, received_at, read_at, deleted, status, type, link_to
			FROM messages WHERE ((from = ? AND to = ?) OR (from = ? AND to = ?)) AND delivered_at <= ? AND delivered_at >= ? LIMIT ?`
	scanner := m.session.Query(query, users[0], users[1], users[1], users[0], timeTo, timeFrom, limit).WithContext(ctx).Iter().Scanner()

	var messages []*model.Message
	for scanner.Next() {
		var message model.Message
		if err := scanner.Scan(&message); err != nil {
			return nil, err
		}

		messages = append(messages, &message)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (m *MessageRepo) GetUserInbox(ctx context.Context, userID string, limit, lastInbox int) ([]*model.UserInbox, error) {
	query := `SELECT user_id, inbox_msg_id, conv_id, conv_msg_id, msg_time, sender, content FROM user_inbox WHERE user_id = ? AND inbox_msg_id > ? LIMIT ?`
	scanner := m.session.Query(query, userID, lastInbox, limit).WithContext(ctx).Iter().Scanner()

	var userInboxes []*model.UserInbox
	for scanner.Next() {
		var userInbox model.UserInbox
		if err := scanner.Scan(&userInbox.UserID,
			&userInbox.InboxMessageID,
			&userInbox.ConversationID,
			&userInbox.ConversationMessageID,
			&userInbox.MessageTime,
			&userInbox.Sender,
			&userInbox.Content); err != nil {
			return nil, err
		}

		userInboxes = append(userInboxes, &userInbox)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return userInboxes, nil
}

func (m *MessageRepo) GetConversationMessages(ctx context.Context, conversationID string, limit int, beforeMsg int64) ([]*model.ConversationMessage, error) {
	query := `SELECT conv_id, conv_msg_id, msg_time, sender, content FROM conv_msg WHERE conv_id = ? AND conv_msg_id < ? LIMIT ?`
	scanner := m.session.Query(query, conversationID, beforeMsg, limit).WithContext(ctx).Iter().Scanner()

	var conversationMessages []*model.ConversationMessage
	for scanner.Next() {
		var conversationMessage model.ConversationMessage
		if err := scanner.Scan(&conversationMessage.ConversationID,
			&conversationMessage.ConversationMessageID,
			&conversationMessage.MessageTime,
			&conversationMessage.Sender,
			&conversationMessage.Content); err != nil {
			return nil, err
		}

		conversationMessages = append(conversationMessages, &conversationMessage)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return conversationMessages, nil
}

func (m *MessageRepo) GetReadReceipt(ctx context.Context, conversationID, userID string) (*model.ReadReceipt, error) {
	query := `SELECT conv_id, user_id, last_seen_msg FROM read_receipt WHERE conv_id = ? AND user_ID = ?`
	var readReceipt model.ReadReceipt
	err := m.session.Query(query, conversationID, userID).WithContext(ctx).Consistency(gocql.One).Scan(&readReceipt.ConversationID, &readReceipt.UserID, &readReceipt.MessageID)
	return &readReceipt, err
}

func (m *MessageRepo) GetReadReceipts(ctx context.Context, conversationID string) ([]*model.ReadReceipt, error) {
	query := `SELECT conv_id, user_id, last_seen_msg FROM read_receipt WHERE conv_id = ?`
	scanner := m.session.Query(query, conversationID).WithContext(ctx).Iter().Scanner()

	var readReceipts []*model.ReadReceipt
	for scanner.Next() {
		var readReceipt model.ReadReceipt
		if err := scanner.Scan(&readReceipt.ConversationID, &readReceipt.UserID, &readReceipt.MessageID); err != nil {
			return nil, err
		}

		readReceipts = append(readReceipts, &readReceipt)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return readReceipts, nil
}

func (m *MessageRepo) UpdateReadReceipts(ctx context.Context, conversationID string, readReceiptUpdate []model.ReadReceiptUpdate) error {
	batch := m.session.NewBatch(gocql.UnloggedBatch).WithContext(ctx)
	query := `UPDATE read_receipt SET last_seen_msg = ? WHERE conv_id = ? AND user_id = ?`
	for _, readReceipt := range readReceiptUpdate {
		batch.Entries = append(batch.Entries, gocql.BatchEntry{
			Stmt:       query,
			Args:       []interface{}{readReceipt.MessageID, conversationID, readReceipt.UserID},
			Idempotent: true,
		})
	}

	err := m.session.ExecuteBatch(batch)
	return err
}

func (m *MessageRepo) CreateConversationMessage(ctx context.Context, conversationID, sender, content string, messageTime int64) (int64, error) {
	getQuery := `SELECT conv_msg_id FROM conv_msg WHERE conv_id = ? LIMIT 1`
	var lastConvMsgID int64
	getErr := m.session.Query(getQuery, conversationID).WithContext(ctx).Scan(&lastConvMsgID)
	if getErr != nil && !errors.Is(getErr, gocql.ErrNotFound) {
		return int64(0), getErr
	}

	createQuery := `INSERT INTO conv_msg(conv_id, conv_msg_id, msg_time, sender, content) VALUES (?, ?, ?, ?, ?)`
	createErr := m.session.Query(createQuery, conversationID, lastConvMsgID+1, messageTime, sender, content).WithContext(ctx).Exec()
	if createErr != nil {
		return int64(0), createErr
	}

	conversationMessage := model.ConversationMessage{
		ConversationID:        conversationID,
		ConversationMessageID: lastConvMsgID + 1,
		MessageTime:           messageTime,
		Sender:                sender,
		Content:               content,
	}

	kafkaMessage := model.KafkaMessage{
		UserID:         sender,
		ConversationID: conversationID,
		Type:           model.MESSAGE_TYPE,
		Timestamp:      messageTime,
		Data:           conversationMessage,
	}
	value, err := json.Marshal(kafkaMessage)
	if err != nil {
		return 0, err
	}

	if err := m.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &m.kafkaTopic, Partition: int32(kafka.PartitionAny)},
		Value:          value,
	}, nil); err != nil {
		return 0, err
	}

	return lastConvMsgID + 1, nil
}

func (m *MessageRepo) InsertUserInbox(ctx context.Context, userID, conversationID, sender, content string, convMsgID, messageTime int64) error {
	getQuery := `SELECT inbox_msg_id FROM user_inbox WHERE user_id = ? LIMIT 1`
	var lastInboxMsgID int64
	getErr := m.session.Query(getQuery, userID).WithContext(ctx).Scan(&lastInboxMsgID)
	if getErr != nil && !errors.Is(getErr, gocql.ErrNotFound) {
		return getErr
	}

	query := `INSERT INTO user_inbox (user_id, inbox_msg_id, conv_id, conv_msg_id, msg_time, sender, content) VALUES (?, ?, ?, ?, ?, ?, ?)`
	err := m.session.Query(query, userID, lastInboxMsgID+1, conversationID, convMsgID, messageTime, sender, content).WithContext(ctx).Exec()
	return err
}

func (m *MessageRepo) DeleteUserInbox(ctx context.Context, userID string) error {
	query := `DELETE FROM user_inbox where user_id = ?`
	err := m.session.Query(query, userID).WithContext(ctx).Exec()
	return err
}
