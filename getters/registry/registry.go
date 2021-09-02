package registry

import (
	"context"
	"fmt"
	"reflect"
)

type IPGetter interface {
	GetIP(ctx context.Context) (string, error)
}

type _Getter struct {
	config interface{}
	new    interface{}
}

var registry = map[string]_Getter{}

func Register(name string, config interface{}, new interface{}) {
	if _, ok := registry[name]; ok {
		panic(fmt.Sprintf(`duplicate getter: %s`, name))
	}
	registry[name] = _Getter{config, new}
}

func Create(t string, unmarshal func(interface{}) error) (IPGetter, error) {
	g, ok := registry[t]
	if !ok {
		return nil, fmt.Errorf(`unknown type: %q`, t)
	}
	typ := reflect.TypeOf(g.config)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	cfg := reflect.New(typ)
	if err := unmarshal(cfg.Interface()); err != nil {
		return nil, fmt.Errorf(`getters: %s: %w`, t, err)
	}
	val := reflect.ValueOf(g.new)
	if val.Type().Kind() != reflect.Func {
		return nil, fmt.Errorf(`getters: %s: not a function`, t)
	}
	if val.Type().NumIn() != 1 || val.Type().NumOut() != 1 {
		return nil, fmt.Errorf(`getters: %s: invalid signature`, t)
	}
	getter := val.Call([]reflect.Value{cfg})[0]
	if !getter.Type().Implements(reflect.TypeOf((*IPGetter)(nil)).Elem()) {
		return nil, fmt.Errorf(`getters: %s: doesn't implement IPGetter`, t)
	}
	gt := getter.Interface().(IPGetter)
	return gt, nil
}
