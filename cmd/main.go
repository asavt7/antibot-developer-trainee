package main

import (
	"github.com/asavt7/antibot-developer-trainee/pkg/configs"
	"github.com/asavt7/antibot-developer-trainee/pkg/server"
	"github.com/asavt7/antibot-developer-trainee/pkg/service"
	"github.com/asavt7/antibot-developer-trainee/pkg/store"
	"log"
	"net/http"
)

func main() {
	conf := configs.NewConfigs()

	inMemStore := store.NewInMemoryStoreRateLimitStore(conf)
	inMemStore.InitStore()

	rateLimitService := service.NewServiceImpl(conf, inMemStore)

	serv := server.NewServer(conf, rateLimitService, http.FileServer(http.Dir("./static")))

	err := serv.RunServer()
	if err != nil {
		log.Fatal(err)
	}
}
