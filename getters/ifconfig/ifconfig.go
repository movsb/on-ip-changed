package ifconfig

import (
	"context"
	"fmt"
	"net"

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

func (i *IfConfig) Get(ctx context.Context) (string, error) {
	face, err := net.InterfaceByName(i.c.Name)
	if err != nil {
		return "", fmt.Errorf("ifconfig: InterfaceByName: %w", err)
	}
	addrs, err := face.Addrs()
	if err != nil {
		return "", fmt.Errorf("ifconfig: Addrs: %w", err)
	}
	if len(addrs) == 0 {
		return ``, fmt.Errorf(`ifconfig: no addrs was found`)
	}
	if i.c.Index < 0 || i.c.Index > len(addrs)-1 {
		return ``, fmt.Errorf(`ifconfig: index out of bound`)
	}
	ipstr := addrs[i.c.Index].String()
	ip := net.ParseIP(ipstr).To4()
	if ip == nil {
		ip2, _, err := net.ParseCIDR(ipstr)
		if err == nil {
			ip = ip2.To4()
		}
	}
	if ip == nil {
		return ``, fmt.Errorf(`ifconfig: %q is not an IPv4 address`, ipstr)
	}
	return ip.String(), nil
}
