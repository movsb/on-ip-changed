package ifconfig

import (
	"context"
	"fmt"
	"net"

	"github.com/movsb/on-ip-changed/utils"
	"github.com/movsb/on-ip-changed/utils/registry"
)

func init() {
	registry.RegisterGetter(`ifconfig`, Config{}, NewIfConfig)
}

type Config struct {
	Name  string `yaml:"name"`
	Index int    `yaml:"index"`
}

type IfConfig struct {
	c *Config
}

func NewIfConfig(c *Config) *IfConfig {
	return &IfConfig{c: c}
}

func (i *IfConfig) Get(ctx context.Context) (utils.IP, error) {
	ipr := utils.IP{}
	face, err := net.InterfaceByName(i.c.Name)
	if err != nil {
		return ipr, fmt.Errorf("ifconfig: InterfaceByName: %w", err)
	}
	addrs, err := face.Addrs()
	if err != nil {
		return ipr, fmt.Errorf("ifconfig: Addrs: %w", err)
	}
	if len(addrs) == 0 {
		return ipr, fmt.Errorf(`ifconfig: no addrs was found`)
	}
	if i.c.Index < 0 || i.c.Index > len(addrs)-1 {
		return ipr, fmt.Errorf(`ifconfig: index out of range`)
	}
	for _, addr := range addrs {
		ip := net.ParseIP(addr.String())
		if ip == nil {
			ip2, _, err := net.ParseCIDR(addr.String())
			if err == nil {
				ip = ip2
			}
		}
		if len(ip) == net.IPv4len || ip.To4() != nil {
			ipr.V4 = ip.To4()
		} else if len(ip) == net.IPv6len {
			ipr.V6 = ip.To16()
		}
	}

	if ipr.V4 == nil {
		return ipr, fmt.Errorf(`ifconfig: no ipv4 address was found`)
	}

	return ipr, nil
}
