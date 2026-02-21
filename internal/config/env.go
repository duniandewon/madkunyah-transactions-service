package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Env struct {
	DatabaseUrl      string
	RedisUrl         string
	Port             string
	JwtSecret        string
	XenditKey        string
	XenditWebhookKey string
}

func getEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Environment variable %s is required", key)
	}
	return val
}

func NewEnv() *Env {
	godotenv.Load()

	return &Env{
		Port:             getEnv("PORT"),
		DatabaseUrl:      getEnv("DATABASE_URL"),
		RedisUrl:         getEnv("REDIS_URL"),
		JwtSecret:        getEnv("JWT_SECRET"),
		XenditKey:        getEnv("XENDIT_SECRET_KEY"),
		XenditWebhookKey: getEnv("XENDIT_WEBHOOK_KEY"),
	}
}
