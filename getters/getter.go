package getters

import (
	"context"
	"fmt"

	_ "github.com/movsb/on-ip-changed/getters/asus"
	_ "github.com/movsb/on-ip-changed/getters/domain"
	_ "github.com/movsb/on-ip-changed/getters/ifconfig"
	_ "github.com/movsb/on-ip-changed/getters/website"
	"github.com/movsb/on-ip-changed/utils"
	"github.com/movsb/on-ip-changed/utils/registry"
)

type Getter interface {
	Get(ctx context.Context) (utils.IP, error)
}

type Unmarshaler struct {
	g Getter
}

func (u *Unmarshaler) Getter() Getter {
	return u.g
}

func (u *Unmarshaler) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var t struct {
		Type string `yaml:"type"`
	}
	if err := unmarshal(&t); err != nil {
		return err
	}
	g, err := registry.CreateGetter(t.Type, unmarshal)
	if err != nil {
		return err
	}
	gg, ok := g.(Getter)
	if !ok {
		return fmt.Errorf("getter: not a getter")
	}
	u.g = gg
	return nil
}
