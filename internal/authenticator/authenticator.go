package authenticator

import (
	"fmt"
	"graduation-thesis/internal/authenticator/handler"
	"graduation-thesis/internal/authenticator/repository"
	"graduation-thesis/internal/authenticator/service"
	"graduation-thesis/pkg/custom_error"
	"graduation-thesis/pkg/storage"
	"net/http"
	"sync"
	"time"

	"github.com/spf13/viper"
)

func Run() {
	viper.SetConfigFile("./config/authenticator/config.yaml")
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

	tokenRepo := repository.NewTokenRepo(redis)
	tokenService := service.NewTokenService(
		tokenRepo,
		viper.GetInt64("at_expires"),
		viper.GetInt64("rt_expires"),
		viper.GetString("access_secret"),
		viper.GetString("refresh_secret"),
		custom_error.MappingError(),
	)
	authService := service.NewAuthService(tokenService, custom_error.MappingError(), viper.GetString("user_service_url"))
	authHandler := handler.NewAuthHandler(authService, tokenService)

	router := handler.GetRouter(authHandler)
	srv := http.Server{
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
		if err := srv.ListenAndServe(); err != nil {
			panic(err)
		}
	}(&wg)

	wg.Wait()
}
