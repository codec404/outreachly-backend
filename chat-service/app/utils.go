package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func getEnv(key string) string {
	return os.Getenv(key)
}

func isProduction() bool {
	return getEnv(AppEnvKey) == AppEnvProduction
}

func loadEnvFile(path string) error {
	return godotenv.Load(path)
}

func readConfigFile(name, fileType, path string) error {
	viper.SetConfigName(name)
	viper.SetConfigType(fileType)
	viper.AddConfigPath(path)
	return viper.ReadInConfig()
}

func unmarshalConfigKey(key string, target interface{}) error {
	return viper.UnmarshalKey(key, target)
}

func NewRootContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}
