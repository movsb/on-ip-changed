package ifconfig

import (
	"context"
	"testing"
)

func TestIfConfig(t *testing.T) {
	t.SkipNow()
	i := NewIfConfig(&Config{
		Name: `tun0`,
	})
	t.Log(i.GetIP(context.TODO()))
}
