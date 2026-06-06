package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pxc1984/flashcards-trainer/backend/api/middleware"
	"github.com/pxc1984/flashcards-trainer/backend/store"
	"github.com/pxc1984/flashcards-trainer/backend/store/interfaces"
)

func main() {
	_ = godotenv.Load()
	InitSettings()
	initLogging()
	storeObj, err := store.InitStore(true, "", "admin", true, "")
	if err != nil {
		panic(err.Error())
	}
	defer func(storeObj interfaces.StoreBase) {
		err := storeObj.Close()
		if err != nil {
			panic(err.Error())
		}
	}(storeObj)

	slog.Debug("loaded settings from .env")

	router := gin.Default()

	origins := make([]string, 0, 2)
	origins = append(origins, "https://*.iamamaev.ru")
	if gin.IsDebugging() {
		origins = append(origins, "http://localhost:5173", fmt.Sprintf("http://localhost:%d", Settings.Port))
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Cookie"},
		AllowCredentials: true,
	}))
	apiGroup := router.Group("/api/v1") // inject here
	apiGroup.Use(middleware.RateLimitMiddleware(middleware.RateLimitCapacity, middleware.RateLimitRefillPerSecond))

	_ = apiGroup.Group("/v1") // inject here

	{
	}

	router.GET("/api/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	addr := fmt.Sprintf("%s:%d", Settings.Host, Settings.Port)
	slog.Info("listening", "addr", addr)

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown:", err)
	}

	slog.Info("server exited")
}
