package dnspod

import (
	"context"

	"github.com/movsb/on-ip-changed/utils"
	"github.com/movsb/on-ip-changed/utils/registry"
)

func init() {
	registry.RegisterHandler(`dnspod`, RenewerConfig{}, NewHandler)
}

type RenewerConfig struct {
	Config `yaml:",inline"`
	Domain string `yaml:"domain"`
	Record string `yaml:"record"`
}

type Handler struct {
	c *RenewerConfig
}

func NewHandler(c *RenewerConfig) *Handler {
	return &Handler{c: c}
}

func (r *Handler) upsert(ctx context.Context, ty string, value string) error {
	d := NewDnsPod(&r.c.Config)
	rec, err := d.FindRecord(ctx, r.c.Domain, ty, r.c.Record)
	if err != nil {
		_, err = d.CreateRecord(ctx, r.c.Domain, r.c.Record, ty, value)
		if err != nil {
			return err
		}
		return nil
	}
	err = d.ModifyRecord(ctx, r.c.Domain, rec.ID, r.c.Record, ty, value)
	if err != nil {
		return err
	}
	return nil
}

func (r *Handler) Handle(ctx context.Context, _, ip utils.IP) error {
	if ip.V4 != nil {
		if err := r.upsert(ctx, `A`, ip.V4.String()); err != nil {
			return err
		}
	}
	if ip.V6 != nil {
		if err := r.upsert(ctx, `AAAA`, ip.V6.String()); err != nil {
			return err
		}
	}
	return nil
}
