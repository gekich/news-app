package main

import (
	"context"
	"fmt"
	"github.com/gekich/news-app/templates"
	"log"
	"net/http"
	"time"

	"github.com/gekich/news-app/config"
	"github.com/gekich/news-app/db"
	"github.com/gekich/news-app/handlers"
	"github.com/gekich/news-app/repository"
	"github.com/gekich/news-app/router"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Mongo.Timeout)*time.Second)
	defer cancel()

	mongoDB, err := db.ConnectMongoDB(cfg.Mongo.URI, ctx)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoDB.Disconnect(ctx)

	postRepo := repository.NewPostRepository(mongoDB.Database(cfg.Mongo.DB))
	postTemplates := templates.PostTemplates()
	postHandler := handlers.NewPostHandler(postRepo, postTemplates, cfg)
	r := router.SetupRouter(postHandler, cfg.App.StaticDirectory)

	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Serving at %s\n", serverAddr)
	log.Printf("http://%s\n", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, r))
}
