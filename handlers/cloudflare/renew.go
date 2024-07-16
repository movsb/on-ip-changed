package cloudflare

import (
	"context"

	"github.com/movsb/on-ip-changed/utils"
	"github.com/movsb/on-ip-changed/utils/registry"
)

func init() {
	registry.RegisterHandler(`cloudflare`, RenewerConfig{}, NewHandler)
}

type RenewerConfig struct {
	Config `yaml:",inline"`
	Name   string `yaml:"name"`
}

type Handler struct {
	c *RenewerConfig
}

func NewHandler(c *RenewerConfig) *Handler {
	return &Handler{c: c}
}

func (r *Handler) Handle(ctx context.Context, _, ip utils.IP) error {
	cf := New(&r.c.Config)
	if ip.V4 != nil {
		if err := cf.Update(`A`, r.c.Name, ip.V4.String()); err != nil {
			return err
		}
	}
	if ip.V6 != nil {
		if err := cf.Update(`AAAA`, r.c.Name, ip.V6.String()); err != nil {
			return err
		}
	}
	return nil
}
