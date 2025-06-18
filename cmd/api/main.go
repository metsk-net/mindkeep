// cmd/api/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/metsk-net/mindkeep/internal/middleware"
	"github.com/metsk-net/mindkeep/pkg/config"
	"github.com/metsk-net/mindkeep/pkg/database"
)

func main() {
	// 環境変数読み込み
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// 設定読み込み
	cfg := config.Load()

	// データベース接続
	db, err := database.NewPostgresDB(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Redis接続
	rdb := database.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword)
	defer rdb.Close()

	// Ginセットアップ
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// ミドルウェア
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.RequestID())

	// CORS設定
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ルーティング
	setupRoutes(r, db, rdb)

	// サーバー起動
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	log.Printf("Server starting on port %s", cfg.Port)

	// シグナル待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}

func setupRoutes(r *gin.Engine, db *database.PostgresDB, rdb *database.RedisClient) {
	// ヘルスチェック
	r.GET("/health", func(c *gin.Context) {
		// DB接続確認
		dbHealthy := db.Ping(c.Request.Context()) == nil
		redisHealthy := rdb.Ping(c.Request.Context()) == nil

		status := "healthy"
		code := http.StatusOK

		if !dbHealthy || !redisHealthy {
			status = "unhealthy"
			code = http.StatusServiceUnavailable
		}

		c.JSON(code, gin.H{
			"status": status,
			"services": gin.H{
				"database": dbHealthy,
				"redis":    redisHealthy,
			},
			"timestamp": time.Now().Unix(),
		})
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		// メモ関連（Day2で実装）
		memos := v1.Group("/memos")
		{
			memos.GET("", func(c *gin.Context) {
				c.JSON(200, gin.H{"memos": []string{}})
			})
		}
	}
}
