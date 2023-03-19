package shell

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/mitchellh/go-homedir"
	"github.com/movsb/on-ip-changed/utils"
	"github.com/movsb/on-ip-changed/utils/registry"
)

func init() {
	registry.RegisterHandler(`shell`, Config{}, NewHandler)
}

type StringOrStringArray struct {
	B  bool // true if S.
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

type Config struct {
	Shell   string              `yaml:"shell"`
	Env     map[string]string   `yaml:"env"`
	Command StringOrStringArray `yaml:"command"`
	WorkDir string              `yaml:"work_dir"`
}

type Handler struct {
	cfg *Config
}

func NewHandler(cfg *Config) *Handler {
	return &Handler{cfg: cfg}
}

func (h *Handler) Handle(ctx context.Context, old, ip utils.IP) error {
	var (
		name string
		args []string
	)
	if h.cfg.Command.B {
		name = h.cfg.Shell
		if name == `` {
			name = `bash`
		}
		args = []string{`-c`, h.cfg.Command.S}
	} else {
		name = h.cfg.Command.SS[0]
		args = h.cfg.Command.SS[1:]
		for i, arg := range args {
			switch arg {
			case `$IP`, `$IPv4`:
				args[i] = ip.V4.String()
			case `$IPv6`:
				args[i] = ip.V6.String()
			case `$OldIP`, `$OldIPv4`:
				args[i] = old.V4.String()
			case `$OldIPv6`:
				args[i] = old.V6.String()
			}
		}
	}

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if h.cfg.WorkDir != `` {
		workDir, err := h.expandHome(h.cfg.WorkDir)
		if err != nil {
			return err
		}
		cmd.Dir = workDir
	}

	for k, v := range h.cfg.Env {
		e := fmt.Sprintf("%s=%s", k, v)
		cmd.Env = append(cmd.Env, e)
	}

	cmd.Env = append(cmd.Env, fmt.Sprintf(`IP=%s`, ip.V4.String()))
	cmd.Env = append(cmd.Env, fmt.Sprintf(`IPv4=%s`, ip.V4.String()))
	cmd.Env = append(cmd.Env, fmt.Sprintf(`IPv6=%s`, ip.V6.String()))
	cmd.Env = append(cmd.Env, fmt.Sprintf(`OldIP=%s`, old.V4.String()))
	cmd.Env = append(cmd.Env, fmt.Sprintf(`OldIPv4=%s`, old.V4.String()))
	cmd.Env = append(cmd.Env, fmt.Sprintf(`OldIPv6=%s`, old.V6.String()))

	if err := cmd.Run(); err != nil {
		log.Printf(`shell: %v`, err)
		return err
	}

	return nil
}

func (h *Handler) expandHome(path string) (string, error) {
	return homedir.Expand(path)
}
