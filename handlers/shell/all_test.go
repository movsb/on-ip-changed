package shell

import (
	"context"
	"net"
	"testing"

	"github.com/movsb/on-ip-changed/utils"
)

func TestShellHandler(t *testing.T) {
	h := NewHandler(&Config{
		Command: StringOrStringArray{B: true, S: `echo $IP`},
	})
	h.Handle(context.Background(), utils.IP{}, utils.IP{V4: net.ParseIP(`1.1.2.2`).To4()})

	h = NewHandler(&Config{
		Command: StringOrStringArray{B: false, SS: []string{`echo`, `$IP`}},
	})
	h.Handle(context.Background(), utils.IP{}, utils.IP{V4: net.ParseIP(`1.1.2.2`).To4()})
}
