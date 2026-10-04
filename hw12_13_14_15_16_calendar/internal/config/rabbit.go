package config

import (
	"errors"
	"fmt"
	"net/url"
)

type RabbitConf struct {
	Host       string `env:"RABBIT_HOST"`
	Port       int    `env:"RABBIT_PORT" env-default:"5672"`
	User       string `env:"RABBIT_USER"`
	Password   string `env:"RABBIT_PASSWORD"`
	VHost      string `env:"RABBIT_VHOST" env-default:"/"`
	Exchange   string `env:"RABBIT_EXCHANGE" env-default:"calendar"`
	Queue      string `env:"RABBIT_QUEUE" env-default:"calendar.notifications"`
	RoutingKey string `env:"RABBIT_ROUTING_KEY" env-default:"notification"`
}

func (r *RabbitConf) URL() (string, error) {
	if r.Host == "" {
		return "", errors.New("RABBIT_HOST is required")
	}
	if r.Port == 0 {
		return "", errors.New("RABBIT_PORT is required")
	}
	if r.User == "" {
		return "", errors.New("RABBIT_USER is required")
	}
	if r.Password == "" {
		return "", errors.New("RABBIT_PASSWORD is required")
	}

	u := &url.URL{
		Scheme: "amqp",
		User:   url.UserPassword(r.User, r.Password),
		Host:   fmt.Sprintf("%s:%d", r.Host, r.Port),
		Path:   r.VHost,
	}

	return u.String(), nil
}
