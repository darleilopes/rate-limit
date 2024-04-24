package ratelimiter

import (
	"net/http"
	"strings"
	"time"

	"github.com/darleilopes/rate-limit/ratelimiter/store"
	"github.com/darleilopes/rate-limit/utils"
	"github.com/spf13/viper"
)

type RateLimiter struct {
	store store.Store
}

func NewRateLimiter(store store.Store) *RateLimiter {
	return &RateLimiter{store: store}
}

func getLimitForKey(key string, isToken bool) (int, int, int) {
	var defaultLimit int
	var defaultExpirationTime int
	var defaultBlockDuration int
	if isToken {
		defaultLimit = utils.GetEnvInt("DEFAULT_TOKEN_REQUEST_LIMIT")
		defaultExpirationTime = utils.GetEnvInt("DEFAULT_TOKEN_EXPIRATION_TIME")
		defaultBlockDuration = utils.GetEnvInt("DEFAULT_TOKEN_BLOCK_DURATION")
	} else {
		defaultLimit = utils.GetEnvInt("DEFAULT_IP_REQUEST_LIMIT")
		defaultExpirationTime = utils.GetEnvInt("DEFAULT_IP_EXPIRATION_TIME")
		defaultBlockDuration = utils.GetEnvInt("DEFAULT_IP_BLOCK_DURATION")
	}

	if isToken {
		tokens, ok := viper.Get("rate_limit.tokens").([]interface{})
		if !ok {
			return defaultLimit, defaultExpirationTime, defaultBlockDuration
		}

		for _, t := range tokens {
			tokenConfig, ok := t.(map[string]interface{})
			if !ok {
				continue
			}
			if tokenConfig["token"].(string) == key {
				limit, ok := tokenConfig["limit"].(int)
				if !ok {
					limit = defaultLimit
				}

				expirationTime, ok := tokenConfig["expiration"].(int)
				if !ok {
					expirationTime = defaultExpirationTime
				}

				blockTime, ok := tokenConfig["block"].(int)
				if !ok {
					blockTime = defaultBlockDuration
				}

				return limit, expirationTime, blockTime
			}
		}
	} else {
		ips, ok := viper.Get("rate_limit.ips").([]interface{})
		if !ok {
			return defaultLimit, defaultExpirationTime, defaultBlockDuration
		}

		for _, ip := range ips {
			ipConfig, ok := ip.(map[string]interface{})
			if !ok {
				continue
			}
			if ipConfig["ip"].(string) == key {
				limit, ok := ipConfig["limit"].(int)
				if !ok {
					limit = defaultLimit
				}

				expirationTime, ok := ipConfig["expiration"].(int)
				if !ok {
					expirationTime = defaultExpirationTime
				}

				blockTime, ok := ipConfig["block"].(int)
				if !ok {
					blockTime = defaultBlockDuration
				}

				return limit, expirationTime, blockTime
			}
		}
	}

	return defaultLimit, defaultExpirationTime, defaultBlockDuration
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		token := r.Header.Get("API_KEY")
		var key string
		var isToken bool

		requesterIPPort := r.Header.Get("X-Forwarded-For")
		if requesterIPPort == "" {
			requesterIPPort = r.RemoteAddr
		}
		requestedIP := strings.Split(requesterIPPort, ",")[0]
		key = requestedIP

		if token != "" {
			key = token
			isToken = true
		} else {
			isToken = false
		}

		blocked, err := rl.store.IsBlocked(ctx, key)
		if blocked {
			http.Error(w, "You have reached the maximum number of requests allowed within a certain time frame (Is Blocked)", http.StatusTooManyRequests)
			return
		}

		limit, expirationTime, blockDuration := getLimitForKey(key, isToken)

		count, err := rl.store.Get(ctx, key)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if count >= limit {
			err := rl.store.Block(ctx, key, time.Duration(blockDuration))
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			http.Error(w, "You have reached the maximum number of requests allowed within a certain time frame (Became Blocked)", http.StatusTooManyRequests)
			return
		}

		_, err = rl.store.Increment(ctx, key, time.Duration(expirationTime))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r)
	})
}
