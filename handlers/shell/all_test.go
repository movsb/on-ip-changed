package shell

import (
	"context"
	"testing"
)

func TestShellHandler(t *testing.T) {
	h := NewHandler(&Config{
		Command: StringOrStringArray{B: true, S: `cat $IP`},
	})
	h.Handle(context.Background(), `1.1.2.2`)

	h = NewHandler(&Config{
		Command: StringOrStringArray{B: false, SS: []string{`cat`, `$IP`}},
	})
	h.Handle(context.Background(), `1.1.2.2`)
}
