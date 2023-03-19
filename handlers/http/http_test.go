package http

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"github.com/movsb/on-ip-changed/utils"
)

func TestHTTP(t *testing.T) {
	ht := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := httputil.DumpRequest(r, true)
		t.Log(string(b))
	}))
	defer ht.Close()
	h := NewHandler(&Config{
		Endpoint: ht.URL,
		Args: map[string]string{
			`A`:  `a`,
			`B`:  `b`,
			`IP`: `$IP`,
		},
		Headers: map[string]string{
			`token`: `ttt`,
		},
	})
	h.Handle(context.Background(), utils.IP{V4: net.IPv4(1, 2, 3, 4)})
}
