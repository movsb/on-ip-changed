package shell

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

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

func (h *Handler) Handle(ctx context.Context, ip string) error {
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
	}

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = h.cfg.WorkDir
	for k, v := range h.cfg.Env {
		e := fmt.Sprintf("%s=%s", k, v)
		cmd.Env = append(cmd.Env, e)
	}

	cmd.Env = append(cmd.Env, fmt.Sprintf(`IP=%s`, ip))

	if err := cmd.Run(); err != nil {
		log.Printf(`shell: %v`, err)
		return err
	}

	return nil
}
