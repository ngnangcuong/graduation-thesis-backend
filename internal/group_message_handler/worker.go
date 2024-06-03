package group_message_handler

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"graduation-thesis/pkg/logger"
	request "graduation-thesis/pkg/requests"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gorilla/websocket"
)

type Worker struct {
	consumer            *kafka.Consumer
	topics              []string
	groupURL            string
	websocketManagerURL string
	maxRetries          int
	timeout             time.Duration
	retryInterval       time.Duration
	pingInterval        time.Duration
	logger              logger.Logger
	mapConnection       *MapConnection
	mapMu               *MapMu
	wg                  *sync.WaitGroup
	Done                chan struct{}
}

func NewWorker(
	consumer *kafka.Consumer,
	topics []string, groupURL,
	websocketManagerURL string,
	timeout time.Duration,
	maxRetries int,
	retryInterval time.Duration,
	pingInterval time.Duration,
	logger logger.Logger) *Worker {
	return &Worker{
		consumer:            consumer,
		topics:              topics,
		groupURL:            groupURL,
		websocketManagerURL: websocketManagerURL,
		maxRetries:          maxRetries,
		timeout:             timeout,
		retryInterval:       retryInterval,
		pingInterval:        pingInterval,
		logger:              logger,
		mapConnection: &MapConnection{
			data: make(map[string]*ChanMessage),
		},
		mapMu: &MapMu{
			data: make(map[string]*sync.Mutex),
		},
		wg:   &sync.WaitGroup{},
		Done: make(chan struct{}),
	}
}

// rebalanceCallback is only useful when we commit manually ?

func (w *Worker) Do() error {
	w.logger.Info("[MAIN] Starting Group Message Handler")
	err := w.consumer.SubscribeTopics(w.topics, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-w.Done:
			w.logger.Info("[MAIN] Shutting down Group Message Handler")
			w.wg.Wait()
			return nil
		default:
			event := w.consumer.Poll(100)
			switch e := event.(type) {
			case *kafka.Message:
				w.logger.Infof("[MAIN] Message at %d[%d]: %v\n", e.TopicPartition.Partition, e.TopicPartition.Offset, e.Value)
				w.processMessage(e)
				w.logger.Infof("[MAIN] Done in processing message at %d[%d]\n", e.TopicPartition.Partition, e.TopicPartition.Offset)
			case kafka.PartitionEOF:
				// TODO: Study of PartitionEOF events'affect --> Normal events, just correnponded for notification
				w.logger.Infof("[MAIN] Reached %v\n", e)
			case kafka.Error:
				// TODO: Error Handling with broker down error
				w.logger.Errorf("[MAIN] Error: %v\n", e)
			default:
			}
		}
	}
}

func (w *Worker) processMessage(message *kafka.Message) {
	var kafkaMessage KafkaMessage
	if err := json.Unmarshal(message.Value, &kafkaMessage); err != nil {
		w.logger.Errorf("[MAIN] Cannot unmarshal message at %d[%d]: %v\n", message.TopicPartition.Partition, message.TopicPartition.Offset, err)
		return
	}

	users, err := w.getConversationUsers(kafkaMessage.ConversationID)
	if err != nil {
		w.logger.Errorf("[MAIN] Failed at get user in coversation %v: %v\n", kafkaMessage.ConversationID, err)
		return
	}

	if len(users) <= 2 {
		w.logger.Info("[MAIN] Ignored: conversation has only two members\n")
		return
	}

	data := kafkaMessage.Data.(map[string]interface{})

	for _, user := range users {
		if user == kafkaMessage.UserID {
			continue
		}
		message := Message{
			ConversationID:        data["conv_id"].(string),
			ConversationMessageID: int64(data["conv_msg_id"].(float64)),
			MessageTime:           int64(data["msg_time"].(float64)),
			Sender:                data["sender"].(string),
			Content:               data["content"].(string),
			Receiver:              user,
		}
		go w.sendMessage(user, message)
	}
}

func (w *Worker) getConversationUsers(conversationID string) ([]string, error) {
	var (
		result interface{}
		err    error
	)
	for i := 1; i <= w.maxRetries; i++ {
		result, err = request.HTTPRequestCall(
			fmt.Sprintf("%s/conversation/%s", w.groupURL, conversationID),
			http.MethodGet,
			"",
			nil,
			w.timeout,
		)
		if err != nil {
			w.logger.Errorf("[MAIN] Failed to get users of conversation %s for %dth time: %v", conversationID, i, err.Error())
			time.Sleep(w.retryInterval)
			continue
		}
		break
	}

	if err != nil {
		return nil, err
	}

	membersInterface := result.(map[string]interface{})["members"].([]interface{})
	members := make([]string, len(membersInterface))
	for i, value := range membersInterface {
		members[i] = fmt.Sprintf("%v", value)
	}
	return members, nil
}

func (w *Worker) getWebsocketHandlerConnectedUser(userID string) (*WebsocketHandler, error) {
	var (
		result interface{}
		err    error
	)
	for i := 1; i <= w.maxRetries; i++ {
		result, err = request.HTTPRequestCall(
			fmt.Sprintf("%s/user/%s", w.websocketManagerURL, userID),
			http.MethodGet,
			"",
			nil,
			w.timeout,
		)
		if err != nil {
			w.logger.Errorf("[MAIN] Failed to get websocket connecting to user %s for %dth time: %v", userID, i, err.Error())
			time.Sleep(w.retryInterval)
			continue
		}
		break
	}

	if err != nil {
		return nil, err
	}
	w.logger.Infof("test %v", result)
	websocketHandler, _ := result.(WebsocketHandler)
	w.logger.Infof("test 1 %v", websocketHandler)
	return &websocketHandler, nil
}

func (w *Worker) sendMessage(userID string, message Message) {
	websocketHandler, err := w.getWebsocketHandlerConnectedUser(userID)
	if err != nil {
		w.logger.Errorf(
			"[MAIN][message_%v_%v] Failed to get websocket connecting to user %s: %v",
			message.ConversationMessageID,
			message.ConversationID,
			userID,
			err.Error(),
		)
		return
	}
	if websocketHandler.ID == "" {
		w.logger.Infof("[MAIN][message_%v_%v] User %v is not online",
			message.ConversationMessageID,
			message.ConversationID,
			userID,
		)
		return
	}

	websocketConnection := w.mapConnection.Get(websocketHandler.ID)
	if websocketConnection != nil && websocketConnection.Send(message) == nil {
		return
	}

	_, eErr := w.establishWebsocketConnection(websocketHandler)
	if eErr != nil {
		w.logger.Errorf(
			"[MAIN][message_%v_%v] Cannot establish a websocket connection to %s at %s: %v",
			message.ConversationMessageID,
			message.ConversationID,
			websocketHandler.ID,
			websocketHandler.IPAddress,
			eErr,
		)
		return
	}
	_ = w.mapConnection.Get(websocketHandler.ID).Send(message)
	return
}

func (w *Worker) establishWebsocketConnection(websocketHandler *WebsocketHandler) (*websocket.Conn, error) {
	websocketHandlerURL := url.URL{Scheme: "ws", Host: websocketHandler.IPAddress, Path: "/peer/ws?websocket_id=group_message_handler_1"}
	conn, _, err := websocket.DefaultDialer.Dial(websocketHandlerURL.String(), nil)
	if err != nil {
		return nil, err
	}
	w.mapConnection.Set(websocketHandler.ID, 100)
	go w.keepWebsocketConnection(conn, websocketHandler.ID)
	return conn, nil
}

func (w *Worker) keepWebsocketConnection(conn *websocket.Conn, websocketHandlerID string) {
	w.wg.Add(1)
	defer w.wg.Done()
	done := make(chan struct{})
	go func(w *Worker, conn *websocket.Conn, done chan struct{}) { // Checking when websocket handler connection is corrupted for any reason
		defer close(done)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				_, isCloseErr := err.(*websocket.CloseError)
				_, isNetErr := err.(*net.OpError)
				if isCloseErr || isNetErr {
					w.mapConnection.Get(websocketHandlerID).Close()
					w.mapConnection.Del(websocketHandlerID)
					w.logger.Infof("[%s][Read] Disconneted connection with websocket handler %s", websocketHandlerID, websocketHandlerID)
					return
				}
				w.logger.Errorf("[%s][Read] Error happens when try to ping to websocket handler %s", websocketHandlerID, websocketHandlerID)
			}
		}
	}(w, conn, done)

	w.logger.Infof("[%s] Starting communicating with websocket handler %s", websocketHandlerID, websocketHandlerID)
	defer conn.Close()
	timer := time.NewTimer(w.pingInterval)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			conn.WriteMessage(websocket.PingMessage, []byte("ping"))
			timer.Reset(w.pingInterval)

		case message, ok := <-w.mapConnection.Get(websocketHandlerID).Channel:
			if !ok { // Channel has been closed
				w.logger.Infof("[%s][Write] Connection to websocket handler %s has already closed\n", websocketHandlerID, websocketHandlerID)
				return
			}
			err := conn.WriteJSON(message)
			if err != nil {
				w.logger.Errorf(
					"[%s][Write] Error while delivering message %v in conversation %v: %v",
					websocketHandlerID, message.ConversationMessageID, message.ConversationID, err,
				)
				continue
			}
			w.logger.Infof("[%s][Write] Message %v in conversation %v is deliveried", websocketHandlerID, message.ConversationMessageID, message.ConversationID)

		case <-done: // Remote connection has closed
			return
		case <-w.Done: // Worker wants to close a current connection
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				w.logger.Errorf("[%s][Write]: Close connection got failed: %v", websocketHandlerID, err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
