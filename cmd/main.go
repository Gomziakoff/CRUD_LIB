package main

import (
	"fmt"
	"net/http"
	"os"

	_ "github.com/Gomziakoff/CRUD_LIB/docs"
	"github.com/Gomziakoff/CRUD_LIB/internal/config"
	"github.com/Gomziakoff/CRUD_LIB/internal/repository/psql"
	"github.com/Gomziakoff/CRUD_LIB/internal/service"
	"github.com/Gomziakoff/CRUD_LIB/internal/transport/rest"
	"github.com/Gomziakoff/CRUD_LIB/pkg/database"
	log "github.com/sirupsen/logrus"
)

const (
	CONFIG_DIR  = "configs"
	CONFIG_FILE = "main"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func main() {
	cfg, err := config.New(CONFIG_DIR, CONFIG_FILE)
	if err != nil {
		log.Fatal(err)
	}

	// init db
	db, err := database.NewPostgresConnection(database.ConnectionInfo{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		Username: cfg.DB.Username,
		DBName:   cfg.DB.Name,
		SSLMode:  cfg.DB.SSLMode,
		Password: cfg.DB.Password,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	booksRepo := psql.NewBooks(db)
	booksService := service.NewBooks(booksRepo)
	handler := rest.NewHandler(booksService)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: handler.InitRouter(),
	}

	log.Info("SERVER STARTED AT")

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
