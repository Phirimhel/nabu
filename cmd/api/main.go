package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	controllers "nabutilivanie/internal/adapters/http/controllers"
	postgresadapter "nabutilivanie/internal/adapters/postgres"
	redisadapter "nabutilivanie/internal/adapters/redis"
	"nabutilivanie/internal/application"
)

func main() {
	ctx := context.Background()
	db, err := pgxpool.New(ctx, requiredEnv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		log.Fatal(err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr(requiredEnv("REDIS_URL"))})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal(err)
	}

	repository := postgresadapter.New(db)
	players := application.NewPlayerService(repository, redisadapter.NewCooldown(rdb))
	targets := application.NewTargetService(repository)
	payments := application.NewPaymentService(repository)
	app := controllers.NewRouter(targets, players, payments, requiredEnv("WEBHOOK_SECRET"))
	srv := &http.Server{Addr: env("HTTP_ADDR", ":8079"), Handler: app, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		log.Printf("listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
func requiredEnv(k string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	log.Fatalf("%s is required", k)
	return ""
}
func env(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
func redisAddr(raw string) string { 
	if len(raw) > len("redis://") && raw[:len("redis://")] == "redis://" {
		raw = raw[len("redis://"):]
		for i, c := range raw {
			if c == '/' {
				return raw[:i]
			}
		}
	}
	return raw
}
