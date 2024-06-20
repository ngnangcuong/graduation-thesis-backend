package group

import (
	"crypto/tls"
	"fmt"
	"graduation-thesis/internal/group/handler"
	"graduation-thesis/internal/group/repository"
	"graduation-thesis/internal/group/service"
	"graduation-thesis/pkg/custom_error"
	"graduation-thesis/pkg/storage"
	"net/http"
	"sync"
	"time"

	"github.com/spf13/viper"
)

func Run() {
	viper.SetConfigFile("./config/group/config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("Not found config file")
		} else {
			panic(err.Error())
		}
	}

	postgre := storage.GetConnectionPool(viper.GetString("postgres.url"))
	defer postgre.Close()
	redis := storage.GetRedisClient(viper.GetString("redis.url"))
	defer redis.Close()
	errorMap := custom_error.MappingError()

	groupRepo := repository.NewGroupRepo(postgre, redis)
	conversationRepo := repository.NewConversationRepo(postgre, redis)

	groupService := service.NewGroupService(postgre, groupRepo, conversationRepo, errorMap)
	conversationService := service.NewConversationService(postgre, conversationRepo, errorMap)

	groupHandler := handler.NewGroupHandler(groupService, viper.GetString("authenticator.url"))
	conversationHandler := handler.NewConversationHandler(conversationService, viper.GetString("authenticator.url"))

	router := handler.GetRouter(groupHandler, conversationHandler)

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
}
