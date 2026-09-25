package config

import (
	"errors"
	"fmt"
	"net/url"
)

type PGConf struct {
	Host     string `env:"DB_HOST"`
	Port     int    `env:"DB_PORT"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	Database string `env:"DB_NAME"`
	SSLMode  string `env:"DB_SSL" env-default:"disable"`
}

func (p *PGConf) DSN() (string, error) {
	if p.Host == "" {
		return "", errors.New("DB_HOST is required")
	}
	if p.Port == 0 {
		return "", errors.New("DB_PORT is required")
	}
	if p.User == "" {
		return "", errors.New("DB_USER is required")
	}
	if p.Database == "" {
		return "", errors.New("DB_NAME is required")
	}
	if p.Password == "" {
		return "", errors.New("DB_PASSWORD is required")
	}

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(p.User, p.Password),
		Host:   fmt.Sprintf("%s:%d", p.Host, p.Port),
		Path:   p.Database,
	}

	if p.SSLMode != "" {
		q := u.Query()
		q.Set("sslmode", p.SSLMode)
		u.RawQuery = q.Encode()
	}

	return u.String(), nil
}
