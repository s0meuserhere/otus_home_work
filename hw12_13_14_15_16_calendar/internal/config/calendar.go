package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type StorageMode int

const (
	ModeMemory StorageMode = 0
	ModeDB     StorageMode = 1
)

type CalendarConf struct {
	Environment string      `env:"ENVIRONMENT" env-default:"local"`
	StorageMode StorageMode `env:"STORAGE_MODE" env-default:"0"`

	Logger LoggerConf
	HTTP   HTTPConf
	DB     PGConf
}

type LoggerConf struct {
	Level string `env:"LOG_LEVEL" env-default:"debug"`
}

type HTTPConf struct {
	Host string `env:"HTTP_HOST" env-default:"0.0.0.0"`
	Port string `env:"HTTP_PORT" env-default:"8080"`
}

func (h HTTPConf) Addr() string {
	return fmt.Sprintf("%s:%s", h.Host, h.Port)
}

func Load(path string) (*CalendarConf, error) {
	cfg := &CalendarConf{}
	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	return cfg, nil
}
