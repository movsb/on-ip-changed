package handlers

import (
	"context"

	"github.com/movsb/on-ip-changed/handlers/dnspod"
	"github.com/movsb/on-ip-changed/handlers/shell"
)

type Config struct {
	Name   string                `yaml:"name"`
	Shell  *shell.Config         `yaml:"shell"`
	DnsPod *dnspod.RenewerConfig `yaml:"dnspod"`
}

type Handler interface {
	Handle(ctx context.Context, ip string) error
}
