package server

import (
	"fmt"
	"github.com/asavt7/antibot-developer-trainee/pkg/configs"
	"github.com/asavt7/antibot-developer-trainee/pkg/service"
	"html/template"
	"log"
	"net/http"
	"time"
)

type Server struct {
	http.Server
	service        *service.Service
	toManyReqTempl template.Template
	config configs.Config
}

func NewServer(config configs.Config, service *service.Service) *Server {
	fs := http.FileServer(http.Dir("./static"))

	mux := http.NewServeMux()

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	s := &Server{
		Server:         *server,
		service:        service,
		config:         config,
	}
	s.initTemplates()

	mux.HandleFunc("/reset", s.resetHandler)
	mux.HandleFunc("/", s.mainHandler(fs))

	return s
}

func (s *Server) RunServer() error {
	log.Printf("Starting server at %s", s.Addr)
	err := s.ListenAndServe()
	return err
}
