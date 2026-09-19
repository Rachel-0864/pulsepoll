package config

import "os"

// Config holds all environment-driven settings for the service.
// Keeping this in one place means main.go and every package that needs
// a setting reads it from here instead of calling os.Getenv all over the code.
type Config struct {
	Port           string
	MongoURI       string
	MongoDBName    string
	RedisAddr      string
	RedisPassword  string
	RedisTLS       bool
	JWTSecret      string
	AllowedOrigin  string
}

func Load() Config {
	return Config{
		Port:          getEnv("PORT", "8080"),
		MongoURI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName:   getEnv("MONGO_DB_NAME", "pollingapp"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisTLS:      getEnv("REDIS_TLS", "false") == "true",
		JWTSecret:     getEnv("JWT_SECRET", "change-me-in-production"),
		AllowedOrigin: getEnv("ALLOWED_ORIGIN", "http://localhost:5173"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
