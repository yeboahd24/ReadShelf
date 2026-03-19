package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpadapter "github.com/dominic/readshelf/internal/adapter/inbound/http"
	"github.com/dominic/readshelf/internal/adapter/inbound/http/handler"
	"github.com/dominic/readshelf/internal/adapter/outbound/claude"
	"github.com/dominic/readshelf/internal/adapter/outbound/cloudinary"
	"github.com/dominic/readshelf/internal/adapter/outbound/groq"
	"github.com/dominic/readshelf/internal/adapter/outbound/nomic"
	"github.com/dominic/readshelf/internal/adapter/outbound/postgres"
	"github.com/dominic/readshelf/internal/adapter/outbound/r2"
	"github.com/dominic/readshelf/internal/config"
	"github.com/dominic/readshelf/internal/core/port/outbound"
	"github.com/dominic/readshelf/internal/core/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Database.
	dbPool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	// Outbound adapters.
	userRepo := postgres.NewUserRepository(dbPool)
	bookRepo := postgres.NewBookRepository(dbPool)
	annotationRepo := postgres.NewAnnotationRepository(dbPool)

	// File storage — swappable via STORAGE_PROVIDER env var.
	var fileStore outbound.FileStore
	switch cfg.StorageProvider {
	case "r2":
		fileStore = r2.NewFileStore(cfg.R2AccountID, cfg.R2AccessKeyID, cfg.R2SecretAccessKey, cfg.R2BucketName)
		log.Println("using Cloudflare R2 for file storage")
	case "cloudinary":
		fileStore = cloudinary.NewFileStore(cfg.CloudinaryCloudName, cfg.CloudinaryAPIKey, cfg.CloudinaryAPISecret)
		log.Println("using Cloudinary for file storage")
	default:
		log.Fatalf("unknown STORAGE_PROVIDER: %s (must be 'r2' or 'cloudinary')", cfg.StorageProvider)
	}

	// Embedder — swappable via EMBEDDER_PROVIDER env var.
	var embedder outbound.Embedder
	switch cfg.EmbedderProvider {
	case "nomic":
		embedder = nomic.NewEmbedder(cfg.HuggingFaceAPIKey)
		log.Println("using Nomic (HuggingFace) for embeddings")
	case "claude":
		embedder = claude.NewEmbedder(cfg.ClaudeAPIKey)
		log.Println("using Claude for embeddings")
	default:
		log.Fatalf("unknown EMBEDDER_PROVIDER: %s (must be 'nomic' or 'claude')", cfg.EmbedderProvider)
	}

	// AI client — swappable via AI_PROVIDER env var.
	var aiClient outbound.AIClient
	switch cfg.AIProvider {
	case "groq":
		aiClient = groq.NewAIClient(cfg.GroqAPIKey)
		log.Println("using Groq (Llama 3.3 70B) for AI recall")
	case "claude":
		aiClient = claude.NewAIClient(cfg.ClaudeAPIKey)
		log.Println("using Claude for AI recall")
	default:
		log.Fatalf("unknown AI_PROVIDER: %s (must be 'groq' or 'claude')", cfg.AIProvider)
	}

	// Services.
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret)
	bookSvc := service.NewBookService(bookRepo, fileStore)
	annotationSvc := service.NewAnnotationService(annotationRepo, bookRepo, embedder)
	recallSvc := service.NewRecallService(annotationRepo, embedder, aiClient)

	// HTTP handlers.
	authHandler := handler.NewAuthHandler(authSvc)
	bookHandler := handler.NewBookHandler(bookSvc)
	annotationHandler := handler.NewAnnotationHandler(annotationSvc)
	recallHandler := handler.NewRecallHandler(recallSvc)

	// Router.
	router := httpadapter.NewRouter(cfg.JWTSecret, authHandler, bookHandler, annotationHandler, recallHandler)

	// Server.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown.
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		log.Println("shutting down server...")
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("server shutdown failed: %v", err)
		}
	}()

	log.Printf("server starting on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
	log.Println("server stopped")
}
