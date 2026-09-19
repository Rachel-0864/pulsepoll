package db

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// ConnectRedis dials Redis. This client is used for two very different
// jobs elsewhere in the app: fast atomic vote counters (HINCRBY) and
// pub/sub fan-out of live updates to WebSocket clients (PUBLISH/SUBSCRIBE).
//
// useTLS should be true for managed providers that require an encrypted
// connection (e.g. Upstash) and false for a plain local Redis instance,
// which doesn't speak TLS by default.
func ConnectRedis(addr, password string, useTLS bool) *redis.Client {
	opts := &redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	}
	if useTLS {
		opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis ping error: %v", err)
	}

	log.Println("connected to Redis")
	return client
}
