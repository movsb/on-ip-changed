package handlers

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/movsb/on-ip-changed/config"
)

type Handler interface {
	Handle(ctx context.Context, ip string)
}

type ShellHandler struct {
	cfg *config.ShellHandlerConfig
}

func NewShellHandler(cfg *config.ShellHandlerConfig) Handler {
	return &ShellHandler{cfg: cfg}
}

func (h *ShellHandler) Handle(ctx context.Context, ip string) {
	var (
		name string
		args []string
	)
	if h.cfg.Command.B {
		name = `bash`
		args = []string{`-c`, h.cfg.Command.S}
	} else {
		name = h.cfg.Command.SS[0]
		args = h.cfg.Command.SS[1:]
	}

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(cmd.Env, fmt.Sprintf(`IP=%s`, ip))

	if err := cmd.Run(); err != nil {
		log.Printf(`shell_handler: %v`, err)
	}
}
