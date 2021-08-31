package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	rootCmd := &cobra.Command{
		Use: filepath.Base(os.Args[0]),
	}

	daemonCmd := &cobra.Command{
		Use:   `daemon`,
		Short: `run daemon`,
		Args:  cobra.NoArgs,
		Run:   daemon,
	}
	daemonCmd.Flags().StringP(`config`, `c`, `config.yaml`, `configuration file`)
	rootCmd.AddCommand(daemonCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func daemon(cmd *cobra.Command, args []string) {
	cfg := readConfig(cmd)

	loop := func() {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Daemon.Timeout)
		defer cancel()
		ip, err := request(ctx, cfg.Sources, cfg.Daemon.Concurrency)
		if err != nil {
			log.Printf(`daemon: error: %v`, err)
			return
		}
		log.Printf(`ip: %s`, ip)
	}

	loop()

	tick := time.NewTicker(cfg.Daemon.Interval)
	defer tick.Stop()

	for range tick.C {
		loop()
	}
}

// TODO: validate
func readConfig(cmd *cobra.Command) *Config {
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
	dec.SetStrict(true)
	if err := dec.Decode(&cfg); err != nil {
		panic(err)
	}
	return &cfg
}
