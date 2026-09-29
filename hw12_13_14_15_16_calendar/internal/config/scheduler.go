package config

import "time"

type SchedulerConf struct {
	Environment string      `env:"ENVIRONMENT" env-default:"local"`
	StorageMode StorageMode `env:"STORAGE_MODE" env-default:"0"`

	Logger    LoggerConf
	DB        PGConf
	Rabbit    RabbitConf
	Scheduler SchedulerParams
}

type SchedulerParams struct {
	// Interval - период запуска планировщика.
	Interval time.Duration `env:"SCHEDULER_INTERVAL" env-default:"1m"`
	// Retention - сколько хранить закончившиеся события.
	Retention time.Duration `env:"SCHEDULER_RETENTION" env-default:"8760h"`
	// RunTimeout - лимит одного запуска, меньше таймаута остановки.
	RunTimeout time.Duration `env:"SCHEDULER_RUN_TIMEOUT" env-default:"5s"`
}

func LoadScheduler(path string) (*SchedulerConf, error) {
	cfg := &SchedulerConf{}
	if err := load(path, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
