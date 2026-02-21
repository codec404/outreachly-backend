package app

import "fmt"

type Config struct {
	Server     ServerConfig
	DB         DBConfig
	JWT        JWTConfig
	CORS       CORSConfig
	RateLimit  RateLimitConfig
	SuperAdmin SuperAdminConfig
}

type ServerConfig struct {
	Host                string `mapstructure:"host"`
	Port                string `mapstructure:"port"`
	MaxRequestBodyBytes int64  `mapstructure:"max_request_body_bytes"`
}

type DBConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
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
}

type SuperAdminConfig struct {
	Name     string
	Email    string
	Password string
}

func Load() (*Config, error) {
	if err := readConfigFile(ConfigFileName, ConfigFileType, ConfigFilePath); err != nil {
		return nil, err
	}

	var server ServerConfig
	if err := unmarshalConfigKey(ViperKeyServer, &server); err != nil {
		return nil, err
	}

	var dbPool DBPoolConfig
	if err := unmarshalConfigKey(ViperKeyDB+".pool", &dbPool); err != nil {
		return nil, err
	}

	var cors CORSConfig
	if err := unmarshalConfigKey(ViperKeyCORS, &cors); err != nil {
		return nil, err
	}

	var rateLimit RateLimitConfig
	if err := unmarshalConfigKey(ViperKeyRateLimit, &rateLimit); err != nil {
		return nil, err
	}

	db := DBConfig{
		Host:     getEnv(EnvDBHost),
		Port:     getEnv(EnvDBPort),
		Name:     getEnv(EnvDBName),
		User:     getEnv(EnvDBUser),
		Password: getEnv(EnvDBPassword),
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

	return &Config{
		Server:     server,
		DB:         db,
		JWT:        jwt,
		CORS:       cors,
		RateLimit:  rateLimit,
		SuperAdmin: superAdmin,
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
	return nil
}
