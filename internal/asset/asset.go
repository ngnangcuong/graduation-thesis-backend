package asset

import (
	"fmt"
	"graduation-thesis/internal/asset/handler"
	"graduation-thesis/internal/asset/repository"
	"graduation-thesis/internal/asset/service"
	"graduation-thesis/pkg/custom_error"
	"graduation-thesis/pkg/logger"
	"net/http"
	"sync"
	"time"

	"github.com/spf13/viper"
)

func Run() {
	viper.SetConfigFile("./config/asset/config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("Not found config file")
		} else {
			panic(err.Error())
		}
	}

	logger, err := logger.GetLogger(
		viper.GetString("logger.level"),
		viper.GetString("logger.path"),
	)
	if err != nil {
		panic(err)
	}

	mapError := custom_error.MappingError()
	assetRepo := repository.NewAssetRepo(
		viper.GetString("seaweed.master_url"),
		viper.GetString("seaweed.volume_url"),
	)
	assetService := service.NewAssetService(assetRepo, mapError, logger)
	assetHandler := handler.NewAssetHandler(viper.GetString("app.local_dir"), assetService)

	router := handler.GetRouter(assetHandler)

	srv := http.Server{
		Addr:         fmt.Sprintf(":%d", viper.GetInt("app.port")),
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 10 * time.Minute,
		IdleTimeout:  1 * time.Minute,
		Handler:      router,
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	serveHTTP := func(wg *sync.WaitGroup) {
		defer wg.Done()
		err := srv.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}

	go serveHTTP(&wg)

	wg.Wait()
}
