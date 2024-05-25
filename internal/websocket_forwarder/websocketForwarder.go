package websocket_forwarder

import (
	"fmt"
	"graduation-thesis/internal/websocket_forwarder/handler"
	"graduation-thesis/pkg/custom_error"
	"graduation-thesis/pkg/logger"
	"net/http"
	"sync"
	"time"

	"github.com/spf13/viper"
)

func Run() {
	viper.SetConfigFile("./config/websocket_forwarder/config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("Not found config file")
		} else {
			panic(err.Error())
		}
	}

	errorMap := custom_error.MappingError()
	logger, err := logger.GetLogger(
		viper.GetString("logger.level"),
		viper.GetString("logger.path"),
	)
	if err != nil {
		panic(err)
	}

	websocketForwarder := handler.NewWebsocketForwarder(
		viper.GetString("websocket_manager.url"),
		errorMap,
		viper.GetDuration("timeout"),
		viper.GetInt("max_retries"),
		viper.GetDuration("retry_interval"),
		logger,
	)
	router := handler.GetRouter(websocketForwarder)

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
	logger.Info("[MAIN] Shutting down Websocket Forwarder")
}
