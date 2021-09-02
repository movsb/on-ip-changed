package config

import (
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// TODO: validate
func ReadConfig(cmd *cobra.Command) *Config {
	configFileString, err := cmd.Flags().GetString(`config`)
	if err != nil {
		panic(err)
	}
	fp, err := os.Open(configFileString)
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
