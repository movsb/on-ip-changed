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
	// if i.c.Index < 0 || i.c.Index > len(addrs)-1 {
	// 	return ipr, fmt.Errorf(`ifconfig: index out of range`)
	// }
	for _, addr := range addrs {
		ip := net.ParseIP(addr.String())
		if ip == nil {
			ip2, _, err := net.ParseCIDR(addr.String())
			if err == nil {
				ip = ip2
			}
		}
		if ip2 := ip.To4(); ip2 != nil {
			ipr.V4 = ip2
		} else if ip2 := ip.To16(); ip2 != nil && ip2.To4() == nil && ip.IsGlobalUnicast() {
			ipr.V6 = ip.To16()
		}
	}

	if ipr.V4 == nil && ipr.V6 == nil {
		return ipr, fmt.Errorf(`ifconfig: no ip v4/v6 address was found`)
	}

	return ipr, nil
}
