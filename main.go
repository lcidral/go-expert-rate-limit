package main

import (
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"

	"go-expert-rater-limit/config"
	"go-expert-rater-limit/limiter"
	"go-expert-rater-limit/middleware"
	"go-expert-rater-limit/storage"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	cfg := config.Load()

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})

	store := storage.NewRedisStorage(redisClient)
	rateLimiter := limiter.NewRateLimiter(store)
	limiterMiddleware := middleware.NewRateLimiterMiddleware(
		rateLimiter,
		cfg.IPLimit,
		cfg.TokenLimit,
		cfg.IPDuration,
		cfg.IPBlockTime,
		cfg.TokenBlockTime,
	)

	mux := http.NewServeMux()
	mux.Handle("/", limiterMiddleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("Heeeey Rater Limit :)"))
		if err != nil {
			log.Println(err)
		}
	})))

	log.Printf("Server starting on port %s", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, mux))
}
