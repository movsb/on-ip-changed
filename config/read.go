package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// TODO: validate
func ReadConfig(path string) *Config {
	fp, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	cfg := Config{}
	dec := yaml.NewDecoder(fp)
	// dec.SetStrict(true)
	if err := dec.Decode(&cfg); err != nil {
		panic(err)
	}
	for _, task := range cfg.Tasks {
		if task.IPv4Only && task.IPv6Only {
			panic(`ipv4only and ipv6only cannot be used together`)
		}
	}
	return &cfg
}
