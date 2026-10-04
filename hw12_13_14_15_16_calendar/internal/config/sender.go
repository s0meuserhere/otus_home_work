package config

type SenderConf struct {
	Environment string `env:"ENVIRONMENT" env-default:"local"`

	Logger LoggerConf
	Rabbit RabbitConf
	Sender SenderParams
}

type SenderParams struct {
	// Prefetch - сколько сообщений брать из очереди без подтверждения.
	Prefetch int `env:"SENDER_PREFETCH" env-default:"10"`
}

func LoadSender(path string) (*SenderConf, error) {
	cfg := &SenderConf{}
	if err := load(path, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
