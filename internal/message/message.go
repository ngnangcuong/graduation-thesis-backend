package message

import (
	"crypto/tls"
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
	messageService := service.NewMessageService(messageRepo, viper.GetString("group_service_url"), logger)
	messageHandler := handler.NewMessageHandler(messageService, viper.GetString("authenticator_url"))

	router := handler.GetRouter(messageHandler)

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
	httpsServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", viper.GetInt("app.https_port")),
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 10 * time.Minute,
		IdleTimeout:  1 * time.Minute,
		Handler:      router,
		TLSConfig:    TLSConfig,
	}

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", viper.GetInt("app.http_port")),
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 10 * time.Minute,
		IdleTimeout:  1 * time.Minute,
		Handler:      router,
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		err := httpsServer.ListenAndServeTLS(viper.GetString("app.cert"), viper.GetString("app.key"))
		if err != nil {
			panic(err.Error())
		}
	}(&wg)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		err := httpServer.ListenAndServe()
		if err != nil {
			panic(err.Error())
		}
	}(&wg)

	wg.Wait()
}
