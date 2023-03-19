package ifconfig

import (
	"context"
	"testing"
)

func TestIfConfig(t *testing.T) {
	i := NewIfConfig(&Config{
		Name: `br0`,
	})
	t.Log(i.Get(context.TODO()))
}
