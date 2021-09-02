package dnspod

import "context"

type RenewerConfig struct {
	Config
	Domain string `yaml:"domain"`
	Record string `yaml:"record"`
}

type Handler struct {
	c *RenewerConfig
}

func NewHandler(c *RenewerConfig) *Handler {
	return &Handler{c: c}
}

func (r *Handler) Handle(ctx context.Context, ip string) error {
	d := NewDnsPod(&r.c.Config)
	rec, err := d.FindRecord(ctx, r.c.Domain, r.c.Record)
	if err != nil {
		_, err = d.CreateRecord(ctx, r.c.Domain, r.c.Record, ip)
		if err != nil {
			return err
		}
		return nil
	}
	err = d.ModifyRecord(ctx, r.c.Domain, rec.ID, r.c.Record, ip)
	if err != nil {
		return err
	}
	return nil
}
