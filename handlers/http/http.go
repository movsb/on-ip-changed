package http

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/movsb/on-ip-changed/utils"
	"github.com/movsb/on-ip-changed/utils/registry"
)

func init() {
	registry.RegisterHandler(`http`, Config{}, NewHandler)
}

type Config struct {
	Endpoint string            `yaml:"endpoint"`
	Args     map[string]string `yaml:"args"`
	Headers  map[string]string `yaml:"headers"`
	Method   string            `yaml:"method"`
	Body     string            `yaml:"body"`
}

type Handler struct {
	cfg *Config
}

func NewHandler(cfg *Config) *Handler {
	return &Handler{cfg: cfg}
}

func (h *Handler) Handle(ctx context.Context, _, ip utils.IP) error {
	method := h.cfg.Method
	if method == "" {
		method = http.MethodGet
	}
	endpoint := h.cfg.Endpoint
	if !strings.Contains(endpoint, `://`) {
		endpoint = `http://` + endpoint
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf(`http: invalid url: %w`, err)
	}
	if len(h.cfg.Args) > 0 {
		args := u.Query()
		for k, v := range h.cfg.Args {
			switch v {
			case `$IP`, `$IPv4`:
				v = ip.V4.String()
			case `$IPv6`:
				v = ip.V6.String()
			}
			args.Set(k, v)
		}
		u.RawQuery = args.Encode()
	}
	var body io.Reader
	if b := h.cfg.Body; len(b) > 0 {
		b = strings.ReplaceAll(b, `$IPv4`, ip.V4.String())
		b = strings.ReplaceAll(b, `$IPv6`, ip.V6.String())
		b = strings.ReplaceAll(b, `$IP`, ip.V4.String())
		body = strings.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return fmt.Errorf(`http: invalid request: %w`, err)
	}
	for k, v := range h.cfg.Headers {
		req.Header.Set(k, v)
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf(`http: error: %w`, err)
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != 200 {
		log.Printf(`http: status code != 200: %d`, rsp.StatusCode)
	}
	return nil
}
