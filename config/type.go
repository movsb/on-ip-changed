package config

import (
	"fmt"
	"time"

	"github.com/movsb/on-ip-changed/getters/domain"
	"github.com/movsb/on-ip-changed/getters/ifconfig"
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
	Name     string               `yaml:"name"`
	Getters  []*GetterUnmarshaler `yaml:"getters"`
	Handlers []*HandlerConfig     `yaml:"handlers"`
}

////////////////////////////////////////////////////////////////

type GetterUnmarshaler struct {
	Website  *WebsiteGetterConfig
	Asus     *AsusGetterConfig
	IfConfig *ifconfig.Config
	Domain   *domain.Config
}

func (s *GetterUnmarshaler) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var t struct {
		Type string `yaml:"type"`
	}
	if err := unmarshal(&t); err != nil {
		return err
	}
	switch t.Type {
	case `website`:
		var w WebsiteGetterConfig
		if err := unmarshal(&w); err != nil {
			return err
		}
		s.Website = &w
		return nil
	case `asus`:
		var a AsusGetterConfig
		if err := unmarshal(&a); err != nil {
			return err
		}
		s.Asus = &a
		return nil
	case `ifconfig`:
		var i ifconfig.Config
		if err := unmarshal(&i); err != nil {
			return err
		}
		s.IfConfig = &i
		return nil
	case `domain`:
		var d domain.Config
		if err := unmarshal(&d); err != nil {
			return err
		}
		s.Domain = &d
		return nil
	default:
		return fmt.Errorf(`unknown type: %q`, t.Type)
	}
}

type WebsiteGetterConfig struct {
	Type   string `yaml:"type"`
	URL    string `yaml:"url"`
	Format string `yaml:"format"`
	Path   string `yaml:"path"`
}

type AsusGetterConfig struct {
	Type     string `yaml:"type"`
	Address  string `yaml:"address"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
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
