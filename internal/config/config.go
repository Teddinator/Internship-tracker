package config

import "time"

type Config struct {
	DBTimeout time.Duration
}

func New() Config {
	return Config{
		DBTimeout: 3 * time.Second,
	}
}
