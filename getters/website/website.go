package website

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/movsb/on-ip-changed/getters/registry"
)

func init() {
	registry.Register(`website`, Config{}, NewWebsite)
}

type Config struct {
	URL    string
	Format string
	Path   string
}

type Website struct {
	c *Config
}

func NewWebsite(c *Config) *Website {
	return &Website{c: c}
}

func (w *Website) GetIP(ctx context.Context) (string, error) {
	return w.roundtrip(ctx, -1, w.c.URL, w.c.Format, w.c.Path)
}

func (w *Website) roundtrip(ctx context.Context, i int, url, format, path string) (string, error) {
	st := time.Now()
	log.Printf(`roundtrip: [%d] sending request to %s`, i, url)
	defer func() {
		et := time.Now()
		log.Printf(`roundtrip: [%d] time elapsed: %v`, i, et.Sub(st))
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ``, fmt.Errorf(`roundtrip: bad request: %w`, err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return ``, fmt.Errorf(`roundtrip: http err: %w`, err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		b, _ := io.ReadAll(io.LimitReader(res.Body, 1<<10))
		return ``, fmt.Errorf(`roundtrip: bad status: %s: %s: %s`, res.Status, url, string(b))
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return ``, fmt.Errorf(`roundtrip: error reading body: %w`, err)
	}
	str := string(body)

	var e Extractor
	switch format {
	case `json`:
		e = NewJsonExtractor(str, path)
	case `raw`:
		e = NewRawExtractor(str)
	case `search`:
		e = NewSearchExtractor(str)
	default:
		return ``, fmt.Errorf(`roundtrip: unknown type: %s`, format)
	}

	ip, err := e.Extract()
	if err != nil {
		return ``, fmt.Errorf(`roundtrip: error extracting: %w`, err)
	}

	return ip, nil
}
