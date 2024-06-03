package worker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"graduation-thesis/internal/websocket_handler/model"
	"graduation-thesis/pkg/custom_error"
	"graduation-thesis/pkg/logger"
	request "graduation-thesis/pkg/requests"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

type Worker struct {
	id                  string
	kafkaProducer       *kafka.Producer
	kafkaTopic          string
	groupServiceUrl     string
	messageServiceUrl   string
	websocketManagerUrl string
	mapUserPeer         *model.MapUserPeer
	mapPeer             *model.MapConnection
	mapUser             *model.MapConnection
	fetchInterval       time.Duration
	pingInterval        time.Duration
	maxRetries          int
	retryInterval       time.Duration
	cacheTimeout        time.Duration
	wg                  *sync.WaitGroup
	logger              logger.Logger
	concurrent          chan struct{}
	done                chan struct{}
}

func NewWorker(
	id string,
	kafkaProducer *kafka.Producer,
	topic string,
	groupServiceUrl string,
	messageServiceUrl string,
	websocketManagerUrl string,
	fetchInterval time.Duration,
	pingInterval time.Duration,
	maxRetries int,
	retryInterval time.Duration,
	cacheTimeout time.Duration,
	logger logger.Logger) *Worker {
	return &Worker{
		id:                  id,
		kafkaProducer:       kafkaProducer,
		kafkaTopic:          topic,
		groupServiceUrl:     groupServiceUrl,
		messageServiceUrl:   messageServiceUrl,
		websocketManagerUrl: websocketManagerUrl,
		mapUserPeer:         model.NewMapUserPeer(),
		mapPeer:             model.NewMapConnection(),
		mapUser:             model.NewMapConnection(),
		fetchInterval:       fetchInterval,
		pingInterval:        pingInterval,
		maxRetries:          maxRetries,
		retryInterval:       retryInterval,
		cacheTimeout:        cacheTimeout,
		wg:                  &sync.WaitGroup{},
		logger:              logger,
		concurrent:          make(chan struct{}, 10000),
		done:                make(chan struct{}),
	}
}

func (w *Worker) removeUserFromMap(connection *model.Connection, userID string) error {
	w.mapUser.Del(userID)
	connection.Delete()
	if err := w.RemoveUser(userID); err != nil {
		return err
	}
	return nil
}

func (w *Worker) removePeerFromMap(connection *model.Connection, websocketID string) {
	w.mapPeer.Del(websocketID)
	connection.Delete()
}

func (w *Worker) EstablishPeerConnetion(websocketHandler *model.WebsocketHandlerClient) error {
	u := url.URL{Scheme: "ws", Host: websocketHandler.IPAddress, Path: fmt.Sprintf("/peer/ws?websocket_id=%s", w.id)}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	go w.KeepPeersConnection(conn, websocketHandler.ID)
	return nil
}

func (w *Worker) KeepPeersConnection(conn *websocket.Conn, websocketID string) error {
	w.logger.Infof("[WebsocketHandler %v] Connected to Websocket Handler %v", websocketID, websocketID)
	defer conn.Close()

	w.wg.Add(1)
	defer w.wg.Done()

	peerConnection := model.Connection{
		WebsocketHandlerID: websocketID,
		WriteChannel:       make(chan model.Message, 100),
		IsDeleted:          false,
	}
	w.mapPeer.Set(websocketID, &peerConnection)

	done := make(chan struct{})
	go func(conn *websocket.Conn, w *Worker, websocketID string, done chan struct{}) {
		defer close(done)
		connection := w.mapPeer.Get(websocketID)
		for {
			var message model.Message
			err := conn.ReadJSON(&message)
			if err != nil {
				_, isCloseErr := err.(*websocket.CloseError)
				_, isNetErr := err.(*net.OpError)
				if isCloseErr || isNetErr { // Connection disconnected
					w.logger.Errorf("[WebsocketHandler %v] Disconnecting to websocket handler %v:%v", websocketID, websocketID, err)
					w.removePeerFromMap(connection, websocketID)
					return
				}
				w.logger.Errorf("[WebsocketHandler %v] Error happens when read message from websocket handler %v:%v", websocketID, websocketID, err)
				continue
			}

			// w.concurrent <- struct{}{}
			go w.ForwardMessage(&message, message.Receiver)
		}
	}(conn, w, websocketID, done)

	connection := w.mapPeer.Get(websocketID)
	for {
		select {
		case message, ok := <-connection.WriteChannel:
			if !ok { // Channel has been closed
				return nil
			}
			err := conn.WriteJSON(message)
			if err != nil {
				_, isCloseErr := err.(*websocket.CloseError)
				_, isNetErr := err.(*net.OpError)
				if isCloseErr || isNetErr { // Connection disconnected
					w.logger.Errorf("[WebsocketHandler %v] Disconnecting to websocket handler %v:%v", websocketID, websocketID, err)
					w.removePeerFromMap(connection, websocketID)

					return nil
				}

				w.logger.Errorf("[WebsocketHandler %v] Error happens when try to send message from websocket handler %v:%v",
					websocketID, websocketID, err)
			}
		case <-done:
			return nil
		case <-w.done:
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				w.logger.Errorf("")
				return err
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return nil
		}
	}

}

func (w *Worker) KeepUsersConnection(conn *websocket.Conn, userID string) error {
	w.logger.Infof("[%v] Connected user %v successfully", userID)
	defer conn.Close()
	if err := w.AddNewUser(userID); err != nil {
		w.logger.Errorf("[%v] Destroying user %v connection because we cannot notify to Websocket Manager: %v", userID, err)
		return err
	}

	w.wg.Add(1)
	defer w.wg.Done()
	userConnection := model.Connection{
		WebsocketHandlerID: w.id,
		WriteChannel:       make(chan model.Message, 100),
		IsDeleted:          false,
	}
	w.mapUser.Set(userID, &userConnection)

	done := make(chan struct{})
	go func(conn *websocket.Conn, w *Worker, userID string, done chan struct{}) { // Read message from user
		defer close(done)
		connection := w.mapUser.Get(userID)
		for {
			var message model.Message
			err := conn.ReadJSON(&message)
			if err != nil {
				_, isCloseErr := err.(*websocket.CloseError)
				_, isNetErr := err.(*net.OpError)
				if isCloseErr || isNetErr { // Connection disconnected
					w.logger.Errorf("[%v] Detroying user %v connection: %v", userID, userID, err)
					if err := w.removeUserFromMap(connection, userID); err != nil {
						w.logger.Errorf("[%v] Removing user %v failed while destroying connection: %v", userID, userID, err)
					}

					return
				}
				w.logger.Errorf("[%v] Error happens when read JSON from user: %v", userID, err.Error())
				continue
			}

			// w.concurrent <- struct{}{}
			w.logger.Debugf("[%v] User %v send message %v", userID, userID, message)
			go w.handleMessageReadFromUser(&message, userID)
		}
	}(conn, w, userID, done)

	go w.ForwardUnreadMessage(conn, userID)

	connection := w.mapUser.Get(userID)
	for {
		select {
		case message, ok := <-connection.WriteChannel:
			if !ok { // Channel has been closed
				return nil
			}
			err := conn.WriteJSON(message)
			if err != nil {
				_, isCloseErr := err.(*websocket.CloseError)
				_, isNetErr := err.(*net.OpError)
				if isCloseErr || isNetErr { // Connection disconnected
					w.logger.Errorf("[%v] Detroying user %v connection: %v", userID, userID, err)
					if err := w.removeUserFromMap(connection, userID); err != nil {
						w.logger.Errorf("[%v] Removing user %v failed while destroying connection: %v", userID, userID, err)
					}

					return nil
				}

				w.logger.Errorf("[%v] Error happens when try to send message to user %v: %v", userID, userID, err)
			}
		case <-done:
			return nil
		case <-w.done:
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				w.logger.Errorf("")
				return err
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return nil
		}
	}
}

func (w *Worker) ForwardUnreadMessage(conn *websocket.Conn, userID string) {
	w.wg.Add(1)
	defer w.wg.Done()

	timer := time.NewTimer(w.fetchInterval)
	defer timer.Stop()

	connection := w.mapUser.Get(userID)
	for {
		select {
		case <-w.done:
			return
		case <-timer.C:
			unreadMessages, err := w.GetUnreadMessage(userID)
			if err != nil {
				w.logger.Errorf("")
				timer.Reset(w.fetchInterval)
				continue
			}

			for _, message := range unreadMessages {
				if !connection.Write(message) {
					return
				}
			}
			timer.Reset(w.fetchInterval)
		}
	}
}

func (w *Worker) handleMessageReadFromUser(message *model.Message, userID string) {
	w.concurrent <- struct{}{}
	defer func() {
		<-w.concurrent
	}()
	var err error

	message.ConversationMessageID, err = w.StoreMessage(message, userID)
	if err != nil {
		w.logger.Errorf("[handleMessageReadFromUser] Cannot store user %v's message: %v", userID, err.Error())
		// <-w.concurrent
		return
	}

	if err := w.ForwardMessage(message, message.Sender); err != nil {
		w.logger.Errorf("[handleMessageReadFromUser] Cannot forward message from user %v to user %v: %v",
			message.Sender, message.Sender, err)
		return
	}

	if message.Receiver == "" {
		// <-w.concurrent
		return
	}

	// users, err := w.GetUsersOfConversation(message.ConversationID)
	// if err != nil {
	// 	w.logger.Errorf("")
	// 	<-w.concurrent
	// 	return
	// }
	// if len(users) > 2 {
	// 	w.logger.Debugf("")
	// 	<-w.concurrent
	// 	return
	// }

	receiveUser := message.Receiver
	if err := w.ForwardMessage(message, receiveUser); err != nil {
		w.logger.Errorf("[handleMessageReadFromUser] Cannot forward message from user %v to user %v: %v",
			message.Sender, receiveUser, err)
		return
	}

	return
}

func (w *Worker) AddNewUser(userID string) error {
	kafkaMessage := model.KafkaMessage{
		WebsocketHandlerID: w.id,
		UserID:             userID,
		Action:             "add",
	}
	value, err := json.Marshal(kafkaMessage)
	if err != nil {
		w.logger.Error("Here")
		return err
	}

	pErr := w.kafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &w.kafkaTopic, Partition: int32(kafka.PartitionAny)},
		Value:          value,
	},
		nil)
	if pErr != nil {
		w.logger.Error("Here1")
		return pErr
	}

	return nil
}

func (w *Worker) RemoveUser(userID string) error {
	kafkaMessage := model.KafkaMessage{
		WebsocketHandlerID: w.id,
		UserID:             userID,
		Action:             "remove",
	}
	value, err := json.Marshal(kafkaMessage)
	if err != nil {
		return err
	}

	pErr := w.kafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &w.kafkaTopic, Partition: int32(kafka.PartitionAny)},
		Value:          value,
	},
		nil)
	if pErr != nil {
		return pErr
	}

	return nil
}

func (w *Worker) StoreMessage(message *model.Message, userID string) (int64, error) {
	sendMessageRequest := model.SendMessageRequest{
		ConversationID: message.ConversationID,
		Sender:         message.Sender,
		Content:        message.Content,
		MessageTime:    message.MessageTime,
	}
	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(&sendMessageRequest); err != nil {
		return int64(0), err
	}

	var (
		result interface{}
		err    error
	)

	for i := 1; i <= w.maxRetries; i++ {
		result, err = request.HTTPRequestCall(
			fmt.Sprintf("%s/message", w.messageServiceUrl),
			http.MethodPost,
			userID,
			body,
			5*time.Second,
		)
		if err != nil {
			w.logger.Errorf("[StoreMessage] Error happen when send message to message service: %v", err.Error())
			time.Sleep(w.retryInterval)
			continue
		}
		break
	}

	if err != nil {
		return int64(0), err
	}

	conversationMessageID, _ := result.(int64)
	return conversationMessageID, nil
}
func (w *Worker) GetUnreadMessage(userID string) ([]model.Message, error) {
	userInbox, err := w.GetUserInbox(userID)
	if err != nil {
		return nil, err
	}

	var messages []model.Message
	for _, inbox := range userInbox {
		message := model.Message{
			ConversationID:        inbox.ConversationID,
			ConversationMessageID: inbox.ConversationMessageID,
			MessageTime:           inbox.MessageTime,
			Sender:                inbox.Sender,
			Content:               inbox.Content,
			Receiver:              userID,
		}
		messages = append(messages, message)
	}
	return messages, nil
}

// func (w *Worker) GetUnreadMessage(userID string) ([]model.Message, error) {
// 	conversations, err := w.GetListConversations(userID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var unreadMessages []model.Message
// 	for _, conversation := range conversations {
// 		readReceipt, rErr := w.GetReadReceipt(conversation, userID)
// 		if rErr != nil {
// 			w.logger.Errorf("")
// 			continue
// 		}
// 		lastMessage, lErr := w.GetLastMessage(conversation)
// 		if lErr != nil {
// 			w.logger.Errorf("")
// 			continue
// 		}

// 		if readReceipt.MessageID < lastMessage.ConversationMessageID {
// 			messages, err := w.GetMessages(conversation, readReceipt.MessageID)
// 			if err != nil {
// 				w.logger.Errorf("")
// 				continue
// 			}
// 			unreadMessages = append(unreadMessages, messages...)
// 		}
// 	}
// 	return unreadMessages, nil
// }

func (w *Worker) GetUsersOfConversation(conversationID string) ([]string, error) {
	var (
		result interface{}
		err    error
	)
	for i := 1; i <= w.maxRetries; i++ {
		result, err = request.HTTPRequestCall(
			fmt.Sprintf("%s/conversation/%s", w.groupServiceUrl, conversationID),
			http.MethodGet,
			"",
			nil,
			5*time.Second,
		)
		if err != nil {
			w.logger.Errorf("")
			time.Sleep(w.retryInterval)
			continue
		}
		break
	}

	if err != nil {
		return nil, err
	}
	users, _ := result.([]string)
	return users, nil
}

func (w *Worker) Register() error {
	ipAddress := GetLocalIP()
	registerRequest := model.AddNewWebsocketHandlerRequest{
		ID:        w.id,
		IPAddress: fmt.Sprintf("%s:%d", ipAddress, viper.GetInt("app.port")),
	}
	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(&registerRequest); err != nil {
		w.logger.Errorf("[Register] Cannot register to Websocket Manager: %v", err)
		return err
	}

	var err error
	for i := 1; i <= w.maxRetries; i++ {
		_, err = request.HTTPRequestCall(
			fmt.Sprintf("%s/websocket_handler/register", w.websocketManagerUrl),
			http.MethodPost,
			"",
			body,
			w.pingInterval,
		)

		if err != nil {
			w.logger.Errorf("[Register] Cannot send register to Websocket Manager for %d times: %v", i, err)
			time.Sleep(w.retryInterval)
			continue
		}
		break
	}
	if err != nil {
		return err
	}

	go w.HeartBeat()
	return nil
}

func (w *Worker) HeartBeat() {
	w.logger.Infof("[Heartbeat] Starting Heartbeat to Websocket Manager")
	timer := time.NewTimer(w.pingInterval)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			pingRequest := model.PingRequest{
				ID:        w.id,
				IPAddress: string(GetLocalIP()),
			}
			body := new(bytes.Buffer)
			if err := json.NewEncoder(body).Encode(&pingRequest); err != nil {
				w.logger.Errorf("[Heartbeat] Cannot encode ping request: %v", err)
				timer.Reset(w.pingInterval)
				continue
			}
			_, err := request.HTTPRequestCall(
				fmt.Sprintf("%s/websocket_handler/ping", w.websocketManagerUrl),
				http.MethodPost,
				"",
				body,
				w.pingInterval,
			)
			if errors.Is(err, custom_error.ErrNotFound) {
				w.logger.Errorf("[Heartbeat] Websocket Manager gets rid of our existence. Register again")
				if err := w.Register(); err != nil {
					w.logger.Errorf("[Heartbeat] Try to Register later because this Register again failed: %v", err)
				}
			}
			w.logger.Debug("[Heartbeat] Sending Heartbeat successfully")
			timer.Reset(w.pingInterval)
		case <-w.done:
			return
		}
	}
}

func (w *Worker) ForwardMessage(message *model.Message, userID string) error {
	w.concurrent <- struct{}{}
	defer func(w *Worker) {
		<-w.concurrent
	}(w)

	// First, websocket handler should check if the user is connecting to itself
	userConnection := w.mapUser.Get(userID)
	if userConnection != nil && userConnection.Write(*message) { // If true, forward message through the connection with the user
		return nil // Make sure that write operator is succeeded
	}

	// If not, check that if the user has been in recent conversation that websocket handler has cached
	peer := w.mapUserPeer.Get(userID)
	if peer != nil { // If so, examine if the cached has expired or not
		websocketID, isExpired := peer.Get()
		if !isExpired { // If the cache is still valueable, we need to ensure that the connection with the peer is maintaining
			peerConnection := w.mapPeer.Get(websocketID)
			if peerConnection != nil && peerConnection.Write(*message) { // When all the thing are true, forward message through that connection
				return nil
			}
		}
	}

	// If any of conditions goes wrong, we conclude that we have no information about this user's connection
	// Then we should query to Websocket Manager to fetch the info
	websocketHandler, err := w.GetWebsocketHandlerConnectUser(userID)
	if err != nil { // Might be connection issue or user has been offline for while
		w.logger.Errorf("[ForwardMessage] Cannot get websocket handler connecting to user %v: %v", userID, err)
		return err
	}
	if websocketHandler.ID == "" {
		w.logger.Errorf("[ForwardMessage] User %v is not online", userID)
		return nil
	}
	if websocketHandler.ID == w.id { // Skip the case that result return is itself
		w.logger.Errorf("[ForwardMessage] Information may not be updated in time")
		return nil
	}

	// We have got a peer connecting to user,
	// so we ask ourself that have we maitained the connection to the peer already or not
	peerConnection := w.mapPeer.Get(websocketHandler.ID)
	if peerConnection == nil { // If not, we're gonna establish the connection with peer
		if err := w.EstablishPeerConnetion(websocketHandler); err != nil { // May be the peer has down ?
			w.logger.Errorf("[ForwardMessage] Cannot establish peer %v connection: %v", websocketHandler.ID, err)
			return err
		}
		time.Sleep(time.Second) // Wait for completing establishing connection
		w.mapUserPeer.Set(userID, model.NewPeer(websocketHandler.ID, w.cacheTimeout))
		w.mapPeer.Get(websocketHandler.ID).Write(*message)
		return nil
	}

	// If we have already, forward this message through
	if peerConnection.Write(*message) {
		w.mapUserPeer.Set(userID, model.NewPeer(websocketHandler.ID, w.cacheTimeout))
	}

	return nil
}

func (w *Worker) GetWebsocketHandlerConnectUser(userID string) (*model.WebsocketHandlerClient, error) {
	var (
		result interface{}
		err    error
	)
	for i := 1; i <= w.maxRetries; i++ {
		result, err = request.HTTPRequestCall(
			fmt.Sprintf("%s/user/%s", w.websocketManagerUrl, userID),
			http.MethodGet,
			"",
			nil,
			5*time.Second,
		)
		if err != nil {
			w.logger.Errorf("[GetWebsocketHandlerConnectUser] Getting Websocket Handler connecting user %v failed for %dth times: %v",
				userID, i, err)
			time.Sleep(w.retryInterval)
			continue
		}
		break
	}

	if err != nil {
		return nil, err
	}
	websocketHandler, _ := result.(model.WebsocketHandlerClient)
	return &websocketHandler, nil
}

func (w *Worker) GetListConversations(userID string) ([]string, error) {
	var (
		result interface{}
		err    error
	)
	for i := 1; i <= w.maxRetries; i++ {
		result, err = request.HTTPRequestCall(
			fmt.Sprintf("%s/conversation/user/%s", w.groupServiceUrl, userID),
			http.MethodGet,
			"",
			nil,
			30*time.Second,
		)
		if err != nil {
			w.logger.Errorf("[GetListConversations] Get list conversations of user %v failed for %dth times: %v",
				userID, i, err)
			continue
		}
		break
	}

	if err != nil {
		return nil, err
	}

	conversations, _ := result.([]string)
	return conversations, nil
}

func (w *Worker) GetLastMessage(conversationID string) (*model.Message, error) {
	var (
		result interface{}
		err    error
	)
	for i := 1; i <= w.maxRetries; i++ {
		result, err = request.HTTPRequestCall(
			fmt.Sprintf("%s/conversation/%s?limit=1", w.messageServiceUrl, conversationID),
			http.MethodGet,
			"",
			nil,
			30*time.Second,
		)
		if err != nil {
			w.logger.Errorf("")
			continue
		}
		break
	}

	if err != nil {
		return nil, err
	}

	message, _ := result.([]model.Message)
	return &message[0], nil
}

func (w *Worker) GetReadReceipt(conversationID, userID string) (*model.ReadReceipt, error) {
	var (
		result interface{}
		err    error
	)

	readReceiptRequest := model.ReadReceiptRequest{
		ConversationID: conversationID,
		UserID:         userID,
	}
	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(&readReceiptRequest); err != nil {
		return nil, err
	}

	for i := 1; i <= w.maxRetries; i++ {
		result, err = request.HTTPRequestCall(
			fmt.Sprintf("%s/read_receipt", w.messageServiceUrl),
			http.MethodPost,
			"",
			body,
			30*time.Second,
		)
		if err != nil {
			w.logger.Errorf("")
			continue
		}
		break
	}

	if err != nil {
		return nil, err
	}

	readReceipt, _ := result.(model.ReadReceipt)
	return &readReceipt, nil
}

func (w *Worker) GetMessages(conversationID string, lastMessageID int64) ([]model.Message, error) {
	var (
		result interface{}
		err    error
	)

	for i := 1; i <= w.maxRetries; i++ {
		result, err = request.HTTPRequestCall(
			fmt.Sprintf("%s/conversation/%s?before_msg=%v", w.messageServiceUrl, conversationID, lastMessageID),
			http.MethodGet,
			"",
			nil,
			5*time.Second,
		)
		if err != nil {
			w.logger.Errorf("")
			continue
		}
		break
	}

	if err != nil {
		return nil, err
	}

	messages, _ := result.([]model.Message)
	return messages, nil
}

func (w *Worker) GetUserInbox(userID string) ([]*model.UserInbox, error) {
	var (
		result interface{}
		err    error
	)

	for i := 1; i <= w.maxRetries; i++ {
		result, err = request.HTTPRequestCall(
			fmt.Sprintf("%s/message/inbox/%s", w.messageServiceUrl, userID),
			http.MethodGet,
			userID,
			nil,
			5*time.Second,
		)
		if err != nil {
			w.logger.Errorf("[GetUserInbox] Cannot get user %v inbox: %v", userID, err)
			continue
		}
		break
	}

	if err != nil {
		return nil, err
	}

	messages, _ := result.([]*model.UserInbox)
	return messages, nil
}

func GetLocalIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil
	}
	defer conn.Close()

	localAddress := conn.LocalAddr().(*net.UDPAddr)

	return localAddress.IP
}

func (w *Worker) Shutdown() {
	w.logger.Info()
	close(w.done)
	w.wg.Wait()
}
