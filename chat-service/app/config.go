package app

import "fmt"

type Config struct {
	Server ServerConfig
	DB     DBConfig
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

type DBConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
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

	db := DBConfig{
		Host:     getEnv(EnvDBHost),
		Port:     getEnv(EnvDBPort),
		Name:     getEnv(EnvDBName),
		User:     getEnv(EnvDBUser),
		Password: getEnv(EnvDBPassword),
	}

	return &Config{Server: server, DB: db}, nil
}

func (c *Config) Validate() error {
	required := []struct{ key, val string }{
		{EnvDBHost, c.DB.Host},
		{EnvDBPort, c.DB.Port},
		{EnvDBName, c.DB.Name},
		{EnvDBUser, c.DB.User},
		{EnvDBPassword, c.DB.Password},
	}
	for _, r := range required {
		if r.val == "" {
			return fmt.Errorf("missing required env var: %s", r.key)
		}
	}
	return nil
}
