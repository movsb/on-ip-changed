package config

import (
	"time"

	"github.com/movsb/on-ip-changed/getters"
	"github.com/movsb/on-ip-changed/handlers"
)

// Config ...
type Config struct {
	Daemon DaemonConfig  `yaml:"daemon"`
	Tasks  []*TaskConfig `yaml:"tasks"`
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Daemon: DefaultDaemonConfig(),
	}
}

////////////////////////////////////////////////////////////////

// DaemonConfig ...
type DaemonConfig struct {
	Interval    time.Duration `yaml:"interval"`
	Concurrency int           `yaml:"concurrency"`
	Timeout     time.Duration `yaml:"timeout"`
	Initial     bool          `yaml:"initial"`
}

// DefaultDaemonConfig ...
func DefaultDaemonConfig() DaemonConfig {
	return DaemonConfig{
		Interval:    time.Minute * 10,
		Concurrency: 5,
		Timeout:     time.Second * 15,
	}
}

////////////////////////////////////////////////////////////////

type TaskConfig struct {
	Name     string                  `yaml:"name"`
	Getters  []*getters.Unmarshaler  `yaml:"getters"`
	Handlers []*handlers.Unmarshaler `yaml:"handlers"`

	IPv4Only bool `yaml:"ipv4only"`
	IPv6Only bool `yaml:"ipv6only"`
}

////////////////////////////////////////////////////////////////
