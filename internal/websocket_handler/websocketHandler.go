package websocket_handler

import (
	"crypto/tls"
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
		viper.GetString("kafka.topic"),
		viper.GetString("3rd_party.group_service_url"),
		viper.GetString("3rd_party.message_service_url"),
		viper.GetString("3rd_party.websocket_manager_url"),
		viper.GetDuration("fetch_interval"),
		viper.GetDuration("ping_interval"),
		viper.GetInt("max_retries"),
		viper.GetDuration("retry_interval"),
		viper.GetDuration("cache_timeout"),
		logger)
	handler := Handler.NewHandler(worker, viper.GetString("3rd_party.authenticator_url"))
	router := Handler.GetRouter(handler)

	if err := worker.Register(); err != nil {
		panic(err)
	}

	TLSConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		},
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true,
	}
	httpsSrv := http.Server{
		Addr:         fmt.Sprintf(":%d", viper.GetInt("app.https_port")),
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 10 * time.Minute,
		IdleTimeout:  1 * time.Minute,
		Handler:      router,
		TLSConfig:    TLSConfig,
	}

	httpSrv := http.Server{
		Addr:         fmt.Sprintf(":%d", viper.GetInt("app.http_port")),
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 10 * time.Minute,
		IdleTimeout:  1 * time.Minute,
		Handler:      router,
	}

	var wg sync.WaitGroup
	wg.Add(2)
	logger.Info("[MAIN] Starting Websocket Forwarder")

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		if err := httpsSrv.ListenAndServeTLS(viper.GetString("app.cert"), viper.GetString("app.key")); err != nil {
			panic(err)
		}
	}(&wg)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		if err := httpSrv.ListenAndServe(); err != nil {
			panic(err)
		}
	}(&wg)

	wg.Wait()
	_ = <-interrupt
	worker.Shutdown()
}
