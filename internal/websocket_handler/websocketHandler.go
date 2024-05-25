package websocket_handler

import (
	"fmt"
	Handler "graduation-thesis/internal/websocket_handler/handler"
	"graduation-thesis/internal/websocket_handler/worker"
	"graduation-thesis/pkg/custom_error"
	"graduation-thesis/pkg/logger"
	"graduation-thesis/pkg/storage"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/spf13/viper"
)

func Run() {
	viper.SetConfigFile("./config/websocket_handler/config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("Not found config file")
		} else {
			panic(err.Error())
		}
	}

	errorMap := custom_error.MappingError()
	_ = errorMap[custom_error.ErrConflict]
	kafkaProducer := storage.NewKafkaProducer(
		viper.GetString("kafka.bootstrap_server"),
		viper.GetInt("kafka.message_max_bytes"),
	)
	defer kafkaProducer.Close()
	logger, err := logger.GetLogger(
		viper.GetString("logger.level"),
		viper.GetString("logger.path"),
	)
	if err != nil {
		panic(err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	worker := worker.NewWorker(
		viper.GetString("id"),
		kafkaProducer,
		viper.GetString("kakfa.topic"),
		viper.GetString("3rd_party.group_service_url"),
		viper.GetString("3rd_party.message_service_url"),
		viper.GetString("3rd_party.websocket_manager_url"),
		viper.GetDuration("fetch_interval"),
		viper.GetDuration("ping_interval"),
		viper.GetInt("max_retries"),
		viper.GetDuration("retry_interval"),
		viper.GetDuration("cache_timeout"),
		logger)
	handler := Handler.NewHandler(worker)
	router := Handler.GetRouter(handler)

	if err := worker.Register(); err != nil {
		panic(err)
	}
	srv := http.Server{
		Addr:         fmt.Sprintf(":%d", viper.GetInt("app.port")),
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 10 * time.Minute,
		IdleTimeout:  1 * time.Minute,
		Handler:      router,
	}

	var wg sync.WaitGroup
	wg.Add(1)
	logger.Info("[MAIN] Starting Websocket Forwarder")

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != nil {
			panic(err)
		}
	}(&wg)

	wg.Wait()
	_ = <-interrupt
	worker.Shutdown()
}
