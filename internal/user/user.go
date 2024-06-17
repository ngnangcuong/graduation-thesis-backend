package user

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"

	"graduation-thesis/internal/user/handler"
	"graduation-thesis/internal/user/repository/token"
	"graduation-thesis/internal/user/repository/user"
	"graduation-thesis/internal/user/service"
	"graduation-thesis/pkg/custom_error"
	"graduation-thesis/pkg/storage"

	"github.com/spf13/viper"
)

func Run() {
	viper.SetConfigFile("./config/user/config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("Not found config file")
		} else {
			panic(err.Error())
		}
	}

	postgres := storage.GetConnectionPool(viper.GetString("postgres.url"))
	defer postgres.Close()

	redisClient := storage.GetRedisClient(viper.GetString("redis.url"))
	defer redisClient.Close()

	userRepoPostgres := user.NewUserRepoPostgres(postgres)
	userRepoRedis := user.NewUserRepoRedis(redisClient)
	tokenRepo := token.NewTokenRepo(redisClient)

	userService := service.NewUserService(postgres, userRepoPostgres, userRepoRedis, custom_error.MappingError())
	tokenService := service.NewTokenService(tokenRepo, viper.GetInt64("token.at_expires"), viper.GetInt64("token.rt_expires"), viper.GetString("token.access_secret"), viper.GetString("token.refresh_secret"))
	authService := service.NewAuthService(userRepoPostgres, tokenService)

	userHandler := handler.NewUserHandler(userService, viper.GetString("authenticator.url"))
	authHandler := handler.NewAuthHandler(authService, tokenService, userService)
	router := handler.GetRouter(authHandler, userHandler)

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
		MinVersion: tls.VersionTLS12,
	}
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", viper.GetInt("app.port")),
		Handler:      router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 10 * time.Minute,
		TLSConfig:    TLSConfig,
	}

	// serveHTTP := func(wg *sync.WaitGroup) {
	// 	defer wg.Done()
	// 	err := srv.ListenAndServe()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }

	serveHTTPS := func(wg *sync.WaitGroup) {
		defer wg.Done()
		err := srv.ListenAndServeTLS(viper.GetString("app.cert"), viper.GetString("app.key"))
		if err != nil {
			panic(err)
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	// go serveHTTP(&wg)
	go serveHTTPS(&wg)
	wg.Wait()
}
