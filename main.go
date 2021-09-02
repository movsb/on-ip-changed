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
	"github.com/movsb/on-ip-changed/getters/registry"
	"github.com/movsb/on-ip-changed/handlers"
	"github.com/movsb/on-ip-changed/handlers/dnspod"
	"github.com/movsb/on-ip-changed/handlers/shell"
	"github.com/spf13/cobra"
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
	cfg := config.ReadConfig(cmd)

	task := func(ctx context.Context, t *config.TaskConfig) {
		last := ``

		log.Printf(`Doing task %s...`, t.Name)
		var gets []registry.IPGetter
		for _, s := range t.Getters {
			gets = append(gets, s.Getter())
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

		for _, hc := range t.Handlers {
			var h handlers.Handler
			switch {
			case hc.Shell != nil:
				h = shell.NewHandler(hc.Shell)
			case hc.DnsPod != nil:
				h = dnspod.NewHandler(hc.DnsPod)
			default:
				log.Println(`unknown handler`)
				continue
			}
			h.Handle(context.Background(), last)
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
