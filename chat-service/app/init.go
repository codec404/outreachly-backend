package app

import log "github.com/codec404/chat-service/pkg/logger"

func Init() (*Config, error) {
	log.InitLogger()

	if !isProduction() {
		loadEnvFile(LocalEnvFile) // soft fail — not present in prod
	}

	cfg, err := Load()
	if err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}
