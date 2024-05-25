package websocket_manager

import (
	"fmt"
	"graduation-thesis/internal/websocket_manager/http_handler"
	"graduation-thesis/internal/websocket_manager/message_consumer"
	"graduation-thesis/internal/websocket_manager/repository"
	"graduation-thesis/internal/websocket_manager/service"
	"graduation-thesis/pkg/custom_error"
	"graduation-thesis/pkg/logger"
	"graduation-thesis/pkg/storage"
	"net/http"
	"sync"
	"time"

	"github.com/spf13/viper"
)

func Run() {
	viper.SetConfigFile("./config/websocket_manager/config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("Not found config file")
		} else {
			panic(err.Error())
		}
	}

	redis := storage.GetRedisClient(viper.GetString("redis.url"))
	defer redis.Close()
	consumer := storage.NewKafkaConsumer(viper.GetString("kafka.bootstrap_servers"), viper.GetString("kafka.group_id"))
	defer consumer.Close()

	errorMap := custom_error.MappingError()
	logger, err := logger.GetLogger(
		viper.GetString("logger.level"),
		viper.GetString("logger.path"),
	)
	if err != nil {
		panic(err)
	}

	userRepo := repository.NewUserRepo(redis)
	websocketManagerRepo := repository.NewWebsocketManagerRepo(redis)

	websocketManagerService := service.NewWebsocketManagerService(
		websocketManagerRepo,
		userRepo,
		errorMap,
		viper.GetInt("service.number_mutex"),
		viper.GetDuration("service.heartbeat_interval"),
		viper.GetInt("service.max_retries"),
		viper.GetDuration("service.retry_interval"),
		logger,
	)
	userService := service.NewUserService(userRepo, errorMap)

	websocketManagerHandler := http_handler.NewWebSocketHandler(websocketManagerService)
	userHandler := http_handler.NewUserHandler(userService)

	messageConsumer := message_consumer.NewMessageConsumer(
		consumer,
		userService,
		websocketManagerService,
		viper.GetInt("service.max_retries"),
		viper.GetDuration("service.retry_interval"),
		logger,
	)
	router := http_handler.GetRouter(userHandler, websocketManagerHandler)

	websocketManagerService.MonitorWebsocketHandler()

	srv := http.Server{
		Addr:         fmt.Sprintf(":%d", viper.GetInt("app.port")),
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 10 * time.Minute,
		IdleTimeout:  1 * time.Minute,
		Handler:      router,
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		if err := messageConsumer.ListenAndServe(viper.GetStringSlice("kafka.topics")); err != nil {
			panic(err)
		}
	}(&wg)

	time.Sleep(time.Second)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != nil {
			panic(err)
		}
	}(&wg)

	wg.Wait()
	logger.Info("[MAIN] Shutting down Websocket Manager")
}
