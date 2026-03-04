package app

import (
	"fmt"

	"github.com/codec404/chat-service/pkg/appconfig"
	log "github.com/codec404/chat-service/pkg/logger"
)

type Config struct {
	Server     ServerConfig
	DB         DBConfig
	JWT        JWTConfig
	CORS       CORSConfig
	RateLimit  RateLimitConfig
	Cleanup    CleanupConfig
	SuperAdmin SuperAdminConfig
	OAuth      OAuthConfig
}

type ServerConfig struct {
	Host                 string `mapstructure:"host"`
	Port                 string `mapstructure:"port"`
	MaxRequestBodyBytes  int64  `mapstructure:"max_request_body_bytes"`
	ShutdownTimeoutSec   int    `mapstructure:"shutdown_timeout_sec"`
	ReadTimeoutSec       int    `mapstructure:"read_timeout_sec"`
	WriteTimeoutSec      int    `mapstructure:"write_timeout_sec"`
	IdleTimeoutSec       int    `mapstructure:"idle_timeout_sec"`
	ReadHeaderTimeoutSec int    `mapstructure:"read_header_timeout_sec"`
}

type DBConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
	Pool     DBPoolConfig
}

type DBPoolConfig struct {
	MaxConns             int32 `mapstructure:"max_conns"`
	MinConns             int32 `mapstructure:"min_conns"`
	MaxConnIdleMinutes   int   `mapstructure:"max_conn_idle_minutes"`
	MaxConnLifetimeHours int   `mapstructure:"max_conn_lifetime_hours"`
}

type JWTConfig struct {
	Secret        string
	AccessExpiry  int // minutes
	RefreshExpiry int // days
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

type RateLimitConfig struct {
	AuthRPM int `mapstructure:"auth_rpm"`
	UserRPM int `mapstructure:"user_rpm"`
}

type CleanupConfig struct {
	TokenCleanupHours int `mapstructure:"token_cleanup_hours"`
}

type SuperAdminConfig struct {
	Name     string
	Email    string
	Password string
}

type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	StateCookieMaxAge  int // seconds; how long the CSRF state cookie lives
}

func Load() (*Config, error) {
	if err := appconfig.Init(ConfigFileName, ConfigFileType, ConfigFilePath); err != nil {
		return nil, err
	}

	var server ServerConfig
	if err := appconfig.UnmarshalKey(ViperKeyServer, &server); err != nil {
		return nil, err
	}

	var dbPool DBPoolConfig
	if err := appconfig.UnmarshalKey(ViperKeyDB+".pool", &dbPool); err != nil {
		return nil, err
	}

	var cors CORSConfig
	if err := appconfig.UnmarshalKey(ViperKeyCORS, &cors); err != nil {
		return nil, err
	}

	var rateLimit RateLimitConfig
	if err := appconfig.UnmarshalKey(ViperKeyRateLimit, &rateLimit); err != nil {
		return nil, err
	}

	var cleanup CleanupConfig
	if err := appconfig.UnmarshalKey(ViperKeyCleanup, &cleanup); err != nil {
		return nil, err
	}

	db := DBConfig{
		Host:     getEnv(EnvDBHost),
		Port:     getEnv(EnvDBPort),
		Name:     getEnv(EnvDBName),
		User:     getEnv(EnvDBUser),
		Password: getEnv(EnvDBPassword),
		SSLMode:  getEnvWithDefault(EnvDBSSLMode, "disable"),
		Pool:     dbPool,
	}

	jwt := JWTConfig{
		Secret:        getEnv(EnvJWTSecret),
		AccessExpiry:  getEnvInt(EnvJWTAccessExpiry, 15),
		RefreshExpiry: getEnvInt(EnvJWTRefreshExpiry, 7),
	}

	superAdmin := SuperAdminConfig{
		Name:     getEnv(EnvSuperAdminName),
		Email:    getEnv(EnvSuperAdminEmail),
		Password: getEnv(EnvSuperAdminPassword),
	}

	var oauthYAML struct {
		StateCookieMaxAge int `mapstructure:"state_cookie_max_age_sec"`
	}
	if err := appconfig.UnmarshalKey(ViperKeyOAuth, &oauthYAML); err != nil {
		return nil, err
	}
	oauth := OAuthConfig{
		GoogleClientID:     getEnv(EnvGoogleClientID),
		GoogleClientSecret: getEnv(EnvGoogleClientSecret),
		GoogleRedirectURL:  getEnv(EnvGoogleRedirectURL),
		StateCookieMaxAge:  oauthYAML.StateCookieMaxAge,
	}

	return &Config{
		Server:     server,
		DB:         db,
		JWT:        jwt,
		CORS:       cors,
		RateLimit:  rateLimit,
		Cleanup:    cleanup,
		SuperAdmin: superAdmin,
		OAuth:      oauth,
	}, nil
}

func (c *Config) Validate() error {
	required := []struct{ key, val string }{
		{EnvDBHost, c.DB.Host},
		{EnvDBPort, c.DB.Port},
		{EnvDBName, c.DB.Name},
		{EnvDBUser, c.DB.User},
		{EnvDBPassword, c.DB.Password},
		{EnvJWTSecret, c.JWT.Secret},
		{EnvSuperAdminName, c.SuperAdmin.Name},
		{EnvSuperAdminEmail, c.SuperAdmin.Email},
		{EnvSuperAdminPassword, c.SuperAdmin.Password},
	}
	for _, r := range required {
		if r.val == "" {
			return fmt.Errorf("missing required env var: %s", r.key)
		}
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("%s must be at least 32 characters (got %d)", EnvJWTSecret, len(c.JWT.Secret))
	}
	if c.Server.MaxRequestBodyBytes <= 0 {
		return fmt.Errorf("server.max_request_body_bytes must be a positive value")
	}
	if isProduction() && c.DB.SSLMode == "disable" {
		log.Warnf("SECURITY WARNING: DB_SSL_MODE=disable in production — database connections are not encrypted")
	}
	return nil
}
