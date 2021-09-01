package domain

import (
	"context"
	"log"
	"testing"
)

func TestDomain(t *testing.T) {
	d := NewDomain(&Config{
		Domain: `home.twofei.com`,
	})
	log.Println(d.GetIP(context.TODO()))
}
