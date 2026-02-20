package app

const (
	// Config file
	ConfigFileName = "config"
	ConfigFileType = "yaml"
	ConfigFilePath = "./configs"

	// Viper keys
	ViperKeyServer = "server"

	// Environment
	AppEnvKey        = "APP_ENV"
	AppEnvProduction = "production"
	LocalEnvFile     = "local.env"

	// Server
	ShutdownTimeout   = 5  // seconds
	ReadTimeout       = 10 // seconds
	WriteTimeout      = 30 // seconds
	IdleTimeout       = 60 // seconds
	ReadHeaderTimeout = 5  // seconds

	// DB env var keys
	EnvDBHost     = "DB_HOST"
	EnvDBPort     = "DB_PORT"
	EnvDBName     = "DB_NAME"
	EnvDBUser     = "DB_USER"
	EnvDBPassword = "DB_PASSWORD"
)
