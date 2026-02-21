// Package appconfig wraps Viper to provide a typed API for reading config.yml values.
package appconfig

import (
	"github.com/spf13/viper"
)

// Init configures Viper and reads the config file.
func Init(name, fileType, path string) error {
	viper.SetConfigName(name)
	viper.SetConfigType(fileType)
	viper.AddConfigPath(path)
	return viper.ReadInConfig()
}

// UnmarshalKey decodes a top-level config key into target (uses mapstructure).
func UnmarshalKey(key string, target any) error {
	return viper.UnmarshalKey(key, target)
}

// GetString returns the string value for key.
func GetString(key string) string {
	return viper.GetString(key)
}

// GetUint returns the uint value for key.
func GetUint(key string) uint {
	return viper.GetUint(key)
}

// GetUint16 returns the uint16 value for key.
func GetUint16(key string) uint16 {
	return viper.GetUint16(key)
}

// GetUint32 returns the uint32 value for key.
func GetUint32(key string) uint32 {
	return viper.GetUint32(key)
}

// GetUint64 returns the uint64 value for key.
func GetUint64(key string) uint64 {
	return viper.GetUint64(key)
}

// GetInt returns the int value for key.
func GetInt(key string) int {
	return viper.GetInt(key)
}

// GetInt64 returns the int64 value for key.
func GetInt64(key string) int64 {
	return viper.GetInt64(key)
}

// GetFloat64 returns the float64 value for key.
func GetFloat64(key string) float64 {
	return viper.GetFloat64(key)
}

// GetBool returns the bool value for key.
func GetBool(key string) bool {
	return viper.GetBool(key)
}

// GetStringSlice returns the string slice value for key.
func GetStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}

// GetIntSlice returns the int slice value for key.
func GetIntSlice(key string) []int {
	return viper.GetIntSlice(key)
}
