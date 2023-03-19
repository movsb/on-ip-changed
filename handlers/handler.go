package handlers

import (
	"context"
	"fmt"

	_ "github.com/movsb/on-ip-changed/handlers/dnspod"
	_ "github.com/movsb/on-ip-changed/handlers/http"
	_ "github.com/movsb/on-ip-changed/handlers/shell"
	"github.com/movsb/on-ip-changed/utils"
	"github.com/movsb/on-ip-changed/utils/registry"
)

type Handler interface {
	Handle(ctx context.Context, ip utils.IP) error
}

type Unmarshaler struct {
	h Handler
}

func (u *Unmarshaler) Handler() Handler {
	return u.h
}

type UnmarshalerHolder func(interface{}) error

func (u *UnmarshalerHolder) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*u = unmarshal
	return nil
}

func (u *Unmarshaler) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var m map[string]UnmarshalerHolder
	if err := unmarshal(&m); err != nil {
		return err
	}
	if len(m) <= 0 || len(m) > 1 {
		return fmt.Errorf(`invalid unmarshaler`)
	}
	var (
		t  string
		uh UnmarshalerHolder
	)
	for k, v := range m {
		t, uh = k, v
		break
	}
	h, err := registry.CreateHandler(t, uh)
	if err != nil {
		return err
	}
	hh, ok := h.(Handler)
	if !ok {
		return fmt.Errorf("handler: not a handler")
	}
	u.h = hh
	return nil
}
