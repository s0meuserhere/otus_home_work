package config

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
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
	GRPC   GRPCConf
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

type GRPCConf struct {
	Host string `env:"GRPC_HOST" env-default:"0.0.0.0"`
	Port string `env:"GRPC_PORT" env-default:"9090"`
}

func (g GRPCConf) Addr() string {
	return fmt.Sprintf("%s:%s", g.Host, g.Port)
}

func LoadCalendar(path string) (*CalendarConf, error) {
	cfg := &CalendarConf{}
	if err := load(path, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// load читает .env-файл, не затирая заданные переменные окружения (cleanenv.ReadConfig их затирает).
// Если файла нет, настройки берутся только из переменных окружения.
func load(path string, cfg any) error {
	if err := godotenv.Load(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("read config file %q: %w", path, err)
	}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	return nil
}
