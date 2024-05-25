package group_message_handler

import (
	"fmt"
	"graduation-thesis/pkg/logger"
	"graduation-thesis/pkg/storage"
	"os"
	"os/signal"
	"sync"

	"github.com/spf13/viper"
)

func Run() {
	viper.SetConfigFile("./config/group_message_handler/config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("Not found config file")
		} else {
			panic(err.Error())
		}
	}

	consumer := storage.NewKafkaConsumer(viper.GetString("kafka.bootstrap_servers"), viper.GetString("kafka.group_id"))
	defer consumer.Close()
	logger, err := logger.GetLogger(
		viper.GetString("logger.level"),
		viper.GetString("logger.path"),
	)
	if err != nil {
		panic(err)
	}

	worker := NewWorker(
		consumer,
		viper.GetStringSlice("kafka.topics"),
		viper.GetString("3rd_party.group_url"),
		viper.GetString("3rd_party.websocket_manager_url"),
		viper.GetDuration("timeout"),
		viper.GetInt("max_retries"),
		viper.GetDuration("retry_interval"),
		viper.GetDuration("ping_interval"),
		logger,
	)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	var wg sync.WaitGroup
	wg.Add(1)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		if err := worker.Do(); err != nil {
			panic(err)
		}
	}(&wg)
	_ = <-interrupt
	close(worker.Done)
	wg.Wait()
}
