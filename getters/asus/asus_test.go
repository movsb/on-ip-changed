package asus

import (
	"context"
	"testing"
)

func TestLogin(t *testing.T) {
	t.SkipNow()
	a := Asus{
		Address:  `192.168.1.1`,
		Username: `asus`,
		Password: `asus`,
	}
	if err := a.login(context.TODO()); err != nil {
		t.Fatal(err)
	}
	t.Logf(`token: %s`, a.token)
	t.Log(a.status(context.TODO()))
}
