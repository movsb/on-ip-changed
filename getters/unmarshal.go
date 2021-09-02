package getters

import (
	_ "github.com/movsb/on-ip-changed/getters/asus"
	_ "github.com/movsb/on-ip-changed/getters/domain"
	_ "github.com/movsb/on-ip-changed/getters/ifconfig"
	"github.com/movsb/on-ip-changed/getters/registry"
	_ "github.com/movsb/on-ip-changed/getters/website"
)

type Unmarshaler struct {
	g registry.IPGetter
}

func (u *Unmarshaler) Getter() registry.IPGetter {
	return u.g
}

func (u *Unmarshaler) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var t struct {
		Type string `yaml:"type"`
	}
	if err := unmarshal(&t); err != nil {
		return err
	}
	g, err := registry.Create(t.Type, unmarshal)
	if err != nil {
		return err
	}
	u.g = g
	return nil
}
