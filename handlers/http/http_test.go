package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"
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
			`A`: `a`,
			`B`: `b`,
		},
		Headers: map[string]string{
			`token`: `ttt`,
		},
	})
	h.Handle(context.Background(), `1.1.2.2`)
}
