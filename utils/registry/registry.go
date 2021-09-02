package registry

import (
	"fmt"
	"reflect"
)

type _Ctor struct {
	config interface{}
	create interface{}
}

var getterRegistry = map[string]_Ctor{}
var handlerRegistry = map[string]_Ctor{}

func RegisterGetter(name string, config interface{}, new interface{}) {
	if _, ok := getterRegistry[name]; ok {
		panic(fmt.Sprintf(`duplicate getter: %s`, name))
	}
	getterRegistry[name] = _Ctor{config, new}
}

func RegisterHandler(name string, config interface{}, new interface{}) {
	if _, ok := handlerRegistry[name]; ok {
		panic(fmt.Sprintf(`duplicate handler: %s`, name))
	}
	handlerRegistry[name] = _Ctor{config, new}
}

func CreateGetter(t string, unmarshal func(interface{}) error) (interface{}, error) {
	return create(getterRegistry, t, unmarshal)
}

func CreateHandler(t string, unmarshal func(interface{}) error) (interface{}, error) {
	return create(handlerRegistry, t, unmarshal)
}

func create(registry map[string]_Ctor, t string, unmarshal func(interface{}) error) (interface{}, error) {
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
	val := reflect.ValueOf(g.create)
	if val.Type().Kind() != reflect.Func {
		return nil, fmt.Errorf(`getters: %s: not a function`, t)
	}
	if val.Type().NumIn() != 1 || val.Type().NumOut() != 1 {
		return nil, fmt.Errorf(`getters: %s: invalid signature`, t)
	}
	return val.Call([]reflect.Value{cfg})[0].Interface(), nil
}
