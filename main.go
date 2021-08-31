package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/movsb/on-ip-changed/config"
	"github.com/movsb/on-ip-changed/getters"
	"github.com/movsb/on-ip-changed/getters/asus"
	"github.com/movsb/on-ip-changed/getters/website"
	"github.com/movsb/on-ip-changed/handlers"
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

	task := func(ctx context.Context, t *config.TaskConfig) {
		last := ``

		log.Printf(`Doing task %s...`, t.Name)
		var gets []getters.IPGetter
		for _, s := range t.Getters {
			switch {
			case s.Asus != nil:
				a := &asus.Asus{
					Address:  s.Asus.Address,
					Username: s.Asus.Username,
					Password: s.Asus.Password,
				}
				gets = append(gets, a)
			case s.Website != nil:
				w := &website.Website{
					URL:    s.Website.URL,
					Format: s.Website.Format,
					Path:   s.Website.Path,
				}
				gets = append(gets, w)
			default:
				panic(`invalid getter`)
			}
		}
		ip, err := getters.Request(ctx, gets, cfg.Daemon.Concurrency)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(ip)

		if last == `` && !cfg.Daemon.Initial {
			last = ip
			return
		}

		last = ip

		for _, h := range t.Handlers {
			switch {
			case h.Shell != nil:
				h := handlers.NewShellHandler(&config.ShellHandlerConfig{
					Command: config.StringOrStringArray{B: true, S: `cat $IP`},
				})
				h.Handle(context.Background(), last)
			default:
				panic(`unknown handler`)
			}
		}
	}

	loop := func() {
		ctx, cancel := context.WithTimeout(context.TODO(), cfg.Daemon.Timeout)
		defer cancel()
		wg := &sync.WaitGroup{}
		for _, t := range cfg.Tasks {
			wg.Add(1)
			go func(t *config.TaskConfig) {
				defer wg.Done()
				task(ctx, t)
			}(t)
		}
		wg.Wait()
	}

	loop()

	tick := time.NewTicker(cfg.Daemon.Interval)
	defer tick.Stop()

	for range tick.C {
		loop()
	}
}

// TODO: validate
func readConfig(cmd *cobra.Command) *config.Config {
	configFileString, err := cmd.Flags().GetString(`config`)
	if err != nil {
		panic(err)
	}
	fp, err := os.Open(configFileString)
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	cfg := config.Config{}
	dec := yaml.NewDecoder(fp)
	// dec.SetStrict(true)
	if err := dec.Decode(&cfg); err != nil {
		panic(err)
	}
	return &cfg
}
