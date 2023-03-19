package cmd

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/movsb/on-ip-changed/config"
	"github.com/movsb/on-ip-changed/getters"
	"github.com/movsb/on-ip-changed/utils"
	"github.com/spf13/cobra"
)

func AddCommands(parent *cobra.Command) {
	daemonCmd := &cobra.Command{
		Use:   `daemon`,
		Short: `run daemon`,
		Args:  cobra.NoArgs,
		Run:   daemon,
	}
	daemonCmd.Flags().StringP(`config`, `c`, `config.yaml`, `configuration file`)
	parent.AddCommand(daemonCmd)
}

func daemon(cmd *cobra.Command, args []string) {
	configFileString, err := cmd.Flags().GetString(`config`)
	if err != nil {
		panic(err)
	}
	cfg := config.ReadConfig(configFileString)

	var tes []*TaskExecutor
	for _, t := range cfg.Tasks {
		log := log.New(os.Stderr, t.Name+`: `, log.LstdFlags|log.Lshortfile)
		te := &TaskExecutor{
			task:        t,
			log:         log,
			initial:     cfg.Daemon.Initial,
			concurrency: cfg.Daemon.Concurrency,
		}
		tes = append(tes, te)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ch := make(chan os.Signal, 1)
		defer close(ch)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		log.Println(`Quiting...`)
		cancel()
	}()

	loop(ctx, &cfg.Daemon, tes)

	if ctx.Err() != nil {
		return
	}

	tick := time.NewTicker(cfg.Daemon.Interval)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			if !errors.Is(ctx.Err(), context.Canceled) {
				log.Println(ctx.Err())
			}
			return
		case <-tick.C:
			log.Println(strings.Repeat(`-`, 80))
			loop(ctx, &cfg.Daemon, tes)
		}
	}
}

func loop(ctx context.Context, daemonConfig *config.DaemonConfig, tes []*TaskExecutor) {
	ctx, cancel := context.WithTimeout(ctx, daemonConfig.Timeout)
	defer cancel()

	wg := &sync.WaitGroup{}
	for _, te := range tes {
		wg.Add(1)
		go func(te *TaskExecutor) {
			defer wg.Done()
			te.Execute(ctx)
		}(te)
	}
	wg.Wait()
}

type TaskExecutor struct {
	task        *config.TaskConfig
	concurrency int
	initial     bool
	ips         utils.IP
	log         *log.Logger
}

func (t *TaskExecutor) Execute(ctx context.Context) {
	t.log.Printf(`executing task %s...`, t.task.Name)
	defer t.log.Printf(`executing task %s... done.`, t.task.Name)
	var gets []getters.Getter
	for _, s := range t.task.Getters {
		gets = append(gets, s.Getter())
	}
	ip, err := getters.Request(ctx, gets, t.concurrency)
	if err != nil {
		t.log.Println(`error: `, err)
		return
	}
	if t.ips.V4 == nil && t.ips.V6 == nil && !t.initial {
		t.ips = ip
		t.log.Printf(`got initial ip: %v`, ip)
		return
	}
	if ip.V4.String() == t.ips.V4.String() && ip.V6.String() == t.ips.V6.String() {
		t.log.Printf(`ip not changed, skip executing handlers`)
		return
	}

	old := t.ips
	t.ips = ip

	for i, hc := range t.task.Handlers {
		h := hc.Handler()
		if err := h.Handle(ctx, old, t.ips); err != nil {
			t.log.Printf(`error executing handler[%d]: %v`, i, err)
			continue
		}
	}
}
