package main

import (
	"fmt"
	"time"
)

// Config ...
type Config struct {
	Daemon   DaemonConfig     `yaml:"daemon"`
	Sources  []*SourceConfig  `yaml:"sources"`
	Handlers []*HandlerConfig `yaml:"handlers"`
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Daemon:   DefaultDaemonConfig(),
		Sources:  nil,
		Handlers: nil,
	}
}

////////////////////////////////////////////////////////////////

// DaemonConfig ...
type DaemonConfig struct {
	Interval    time.Duration `yaml:"interval"`
	Concurrency int           `yaml:"concurrency"`
	Timeout     time.Duration `yaml:"timeout"`
}

// DefaultDaemonConfig ...
func DefaultDaemonConfig() DaemonConfig {
	return DaemonConfig{
		Interval:    time.Minute * 10,
		Concurrency: 5,
		Timeout:     time.Second * 15,
	}
}

// SourceConfig ...
type SourceConfig struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

////////////////////////////////////////////////////////////////

type StringOrStringArray struct {
	B  bool
	S  string
	SS []string
}

func (s *StringOrStringArray) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal(&s.S); err == nil {
		s.B = true
		return nil
	}
	if err := unmarshal(&s.SS); err == nil {
		s.B = false
		return nil
	}
	return fmt.Errorf(`expect string or string array`)
}

// HandlerConfig ...
type HandlerConfig struct {
	Name  string              `yaml:"name"`
	Shell *ShellHandlerConfig `yaml:"shell"`
}

type ShellHandlerConfig struct {
	Command StringOrStringArray `yaml:"command"`
}
