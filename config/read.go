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
	return &cfg
}
