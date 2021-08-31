package handlers

import (
	"context"
	"testing"

	"github.com/movsb/on-ip-changed/config"
)

func TestShellHandler(t *testing.T) {
	h := NewShellHandler(&config.ShellHandlerConfig{
		Command: config.StringOrStringArray{B: true, S: `cat $IP`},
	})
	h.Handle(context.Background(), `1.1.2.2`)
}
