package app

import (
	"os"

	"github.com/codec404/chat-service/utils"
	"github.com/joho/godotenv"
)

func getEnv(key string) string {
	return os.Getenv(key)
}

func getEnvWithDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func isProduction() bool {
	return getEnv(AppEnvKey) == AppEnvProduction
}

func loadEnvFile(path string) error {
	return godotenv.Load(path)
}

func getEnvInt(key string, defaultVal int) int {
	if v := getEnv(key); v != "" {
		if n, err := utils.FormatInt(v); err == nil {
			return n
		}
	}
	return defaultVal
}
