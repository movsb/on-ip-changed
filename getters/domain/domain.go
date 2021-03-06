package domain

import (
	"context"
	"fmt"
	"net"

	"github.com/movsb/on-ip-changed/utils/registry"
)

func init() {
	registry.RegisterGetter(`domain`, Config{}, NewDomain)
}

type Config struct {
	Domain string `yaml:"domain"`
}

type Domain struct {
	c *Config
}

func NewDomain(c *Config) *Domain {
	return &Domain{c: c}
}

func (d *Domain) Get(ctx context.Context) (string, error) {
	r := &net.Resolver{}
	ips, err := r.LookupIPAddr(ctx, d.c.Domain)
	if err != nil {
		return "", fmt.Errorf("ifconfig: resolve: %w", err)
	}
	for _, ip := range ips {
		if ip2 := ip.IP.To4(); ip2 != nil {
			return ip2.String(), nil
		}
	}
	return "", fmt.Errorf("ifconfig: no ipv4 address")
}
