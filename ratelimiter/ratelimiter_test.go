package ratelimiter

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	ratelimiter "github.com/darleilopes/rate-limit/ratelimiter/store/redis"
	"github.com/go-redis/redismock/v8"
)

func setupMockRateLimiter() (*RateLimiter, *httptest.Server, redismock.ClientMock) {
	os.Setenv("DEFAULT_IP_EXPIRATION_TIME", "10")
	os.Setenv("DEFAULT_TOKEN_EXPIRATION_TIME", "5")
	os.Setenv("DEFAULT_IP_REQUEST_LIMIT", "1")
	os.Setenv("DEFAULT_TOKEN_REQUEST_LIMIT", "2")
	os.Setenv("DEFAULT_IP_BLOCK_DURATION", "10")
	os.Setenv("DEFAULT_TOKEN_BLOCK_DURATION", "5")

	db, mock := redismock.NewClientMock()

	store := ratelimiter.NewRedisStore(db)
	rateLimiter := NewRateLimiter(store)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	testServer := httptest.NewServer(rateLimiter.Limit(handler))

	return rateLimiter, testServer, mock
}

func UnsetEnvs() {
	os.Unsetenv("DEFAULT_IP_EXPIRATION_TIME")
	os.Unsetenv("DEFAULT_TOKEN_EXPIRATION_TIME")
	os.Unsetenv("DEFAULT_IP_REQUEST_LIMIT")
	os.Unsetenv("DEFAULT_TOKEN_REQUEST_LIMIT")
	os.Unsetenv("DEFAULT_IP_BLOCK_DURATION")
	os.Unsetenv("DEFAULT_TOKEN_BLOCK_DURATION")
}

func TestRateLimiterUnderLoadIP(t *testing.T) {
	_, testServer, mock := setupMockRateLimiter()
	defer testServer.Close()

	expectedKey := "192.168.1.1"

	blockedKey := expectedKey + ":blocked"

	mock.ExpectGet(blockedKey).RedisNil()
	mock.ExpectGet(expectedKey).SetVal("0")
	mock.ExpectIncr(expectedKey).SetVal(1)
	mock.ExpectGet(blockedKey).RedisNil()
	mock.ExpectGet(expectedKey).SetVal("1")
	mock.ExpectIncr(blockedKey).SetVal(1)
	mock.ExpectGet(blockedKey).SetVal("1")
	mock.ExpectGet(blockedKey).SetVal("1")
	mock.ExpectGet(blockedKey).SetVal("1")

	for i := 0; i < 5; i++ {
		req, err := http.NewRequest("GET", testServer.URL, nil)
		if err != nil {
			UnsetEnvs()
			t.Fatal(err)
		}
		req.Header.Set("X-Forwarded-For", expectedKey)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			UnsetEnvs()
			t.Fatalf("Failed to make request: %v", err)
		}
		resp.Body.Close()

		if i >= 2 && resp.StatusCode != http.StatusTooManyRequests {
			UnsetEnvs()
			t.Errorf("Expected HTTP 429 Too Many Requests on attempt %d, got %d", i+1, resp.StatusCode)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		UnsetEnvs()
		t.Errorf("There were unmet expectations: %s", err)
	}

	UnsetEnvs()
}

func TestRateLimiterUnderLoadToken(t *testing.T) {
	_, testServer, mock := setupMockRateLimiter()
	defer testServer.Close()

	expectedKey := "tokenPotato"
	blockedKey := expectedKey + ":blocked"

	mock.ExpectGet(blockedKey).RedisNil()
	mock.ExpectGet(expectedKey).SetVal("0")
	mock.ExpectIncr(expectedKey).SetVal(1)
	mock.ExpectGet(blockedKey).RedisNil()
	mock.ExpectGet(expectedKey).SetVal("1")
	mock.ExpectIncr(expectedKey).SetVal(2)
	mock.ExpectGet(blockedKey).RedisNil()
	mock.ExpectGet(expectedKey).SetVal("2")
	mock.ExpectIncr(blockedKey).SetVal(1)
	mock.ExpectGet(blockedKey).SetVal("1")
	mock.ExpectGet(blockedKey).SetVal("1")

	for i := 0; i < 5; i++ {
		req, err := http.NewRequest("GET", testServer.URL, nil)
		if err != nil {
			UnsetEnvs()
			t.Fatal(err)
		}
		req.Header.Set("API_KEY", expectedKey)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			UnsetEnvs()
			t.Fatalf("Failed to make request: %v", err)
		}
		resp.Body.Close()

		if i >= 2 && resp.StatusCode != http.StatusTooManyRequests {
			UnsetEnvs()
			t.Errorf("Expected HTTP 429 Too Many Requests on attempt %d, got %d", i+1, resp.StatusCode)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		UnsetEnvs()
		t.Errorf("There were unmet expectations: %s", err)
	}

	UnsetEnvs()
}
