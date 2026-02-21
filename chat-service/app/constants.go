package app

const (
	// Config file
	ConfigFileName = "config"
	ConfigFileType = "yaml"
	ConfigFilePath = "./configs"

	// Viper keys (config.yml sections)
	ViperKeyServer    = "server"
	ViperKeyDB        = "db"
	ViperKeyCORS      = "cors"
	ViperKeyRateLimit = "rate_limit"

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

	// Super admin seed env var keys
	EnvSuperAdminName     = "SUPER_ADMIN_NAME"
	EnvSuperAdminEmail    = "SUPER_ADMIN_EMAIL"
	EnvSuperAdminPassword = "SUPER_ADMIN_PASSWORD"

	// JWT env var keys
	EnvJWTSecret        = "JWT_SECRET"
	EnvJWTAccessExpiry  = "JWT_ACCESS_EXPIRY_MINUTES"
	EnvJWTRefreshExpiry = "JWT_REFRESH_EXPIRY_DAYS"
)
