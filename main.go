package main

import (
	"log"
	"net/http"

	"github.com/darleilopes/rate-limit/ratelimiter"
	redisStore "github.com/darleilopes/rate-limit/ratelimiter/store/redis"
	"github.com/spf13/viper"
)

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
}

func main() {
	initConfig()

	store := redisStore.NewRedisStore(nil)

	rateLimiter := ratelimiter.NewRateLimiter(store)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Limites limitado - LIMITRATE"))
	})

	wrappedMux := rateLimiter.Limit(mux)

	http.ListenAndServe(":8080", wrappedMux)
}
