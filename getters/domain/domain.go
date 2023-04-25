package domain

import (
	"context"
	"fmt"
	"net"

	"github.com/movsb/on-ip-changed/utils"
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

func (d *Domain) Get(ctx context.Context) (utils.IP, error) {
	r := &net.Resolver{}
	ipr := utils.IP{}
	ips, err := r.LookupIPAddr(ctx, d.c.Domain)
	if err != nil {
		return ipr, fmt.Errorf("ifconfig: resolve: %w", err)
	}
	for _, ip := range ips {
		if ip2 := ip.IP.To4(); ip2 != nil {
			ipr.V4 = ip2
		}
		if ip2 := ip.IP.To16(); ip2 != nil && ip2.To4() == nil && ip2.IsGlobalUnicast() {
			ipr.V6 = ip2
		}
	}
	if ipr.V4 != nil || ipr.V6 != nil {
		return ipr, nil
	}

	return ipr, fmt.Errorf("ifconfig: no ip v4/v6 addresses")
}
