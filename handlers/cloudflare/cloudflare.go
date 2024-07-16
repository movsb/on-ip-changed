package cloudflare

import (
	"context"
	"fmt"

	cf "github.com/cloudflare/cloudflare-go"
)

type Config struct {
	Token  string `yaml:"token"`
	ZoneID string `yaml:"zone_id"`
}

type Cloudflare struct {
	cfg *Config
	api *cf.API
}

func New(c *Config) *Cloudflare {
	api, err := cf.NewWithAPIToken(c.Token)
	if err != nil {
		panic(err)
	}
	return &Cloudflare{
		cfg: c,
		api: api,
	}
}

func (c *Cloudflare) Update(ty, name, content string) error {
	list, _, err := c.api.ListDNSRecords(context.Background(), cf.ZoneIdentifier(c.cfg.ZoneID), cf.ListDNSRecordsParams{
		Type: ty,
		Name: name,
	})
	if err != nil {
		return err
	}
	if len(list) <= 0 {
		return fmt.Errorf(`dns 记录未找到`)
	} else if len(list) != 1 {
		return fmt.Errorf(`返回的 dns 记录太多`)
	}
	r := list[0]
	_, err = c.api.UpdateDNSRecord(context.Background(), cf.ZoneIdentifier(c.cfg.ZoneID), cf.UpdateDNSRecordParams{
		Type:     r.Type,
		Name:     r.Name,
		Content:  content,
		Data:     r.Data,
		ID:       r.ID,
		Priority: r.Priority,
		TTL:      r.TTL,
		Proxied:  r.Proxied,
		Comment:  &r.Comment,
		Tags:     r.Tags,
	})
	if err != nil {
		return err
	}
	return nil
}
