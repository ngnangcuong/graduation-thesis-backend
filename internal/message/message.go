package message

import (
	"fmt"
	"graduation-thesis/internal/message/handler"
	"graduation-thesis/internal/message/repository"
	"graduation-thesis/internal/message/service"
	"graduation-thesis/pkg/logger"
	"graduation-thesis/pkg/storage"
	"net/http"
	"sync"
	"time"

	"github.com/spf13/viper"
)

func Run() {
	viper.SetConfigFile("./config/message/config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("Not found config file")
		} else {
			panic(err.Error())
		}
	}

	session := storage.GetSession(viper.GetStringSlice("cassandra.hosts"), viper.GetString("cassandra.keyspace"))
	defer session.Close()

	kafkaProducer := storage.GetKafkaProducer(viper.GetString("kafka.bootstrap_servers"), viper.GetInt("kafka.message_max_bytes"))
	defer kafkaProducer.Close()

	logger, err := logger.GetLogger(
		viper.GetString("logger.level"),
		viper.GetString("logger.path"),
	)
	if err != nil {
		panic(err)
	}

	messageRepo := repository.NewMessageRepo(session, kafkaProducer, viper.GetString("kafka.topic"))
	messageService := service.NewMessageService(messageRepo, logger)
	messageHandler := handler.NewMessageHandler(messageService)

	router := handler.GetRouter(messageHandler)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", viper.GetInt("app.port")),
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 10 * time.Minute,
		IdleTimeout:  1 * time.Minute,
		Handler:      router,
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		err := server.ListenAndServe()
		if err != nil {
			panic(err.Error())
		}
	}(&wg)

	wg.Wait()
}
