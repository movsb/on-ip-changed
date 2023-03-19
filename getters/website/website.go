package website

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/movsb/on-ip-changed/utils"
	"github.com/movsb/on-ip-changed/utils/registry"
)

func init() {
	registry.RegisterGetter(`website`, Config{}, NewWebsite)
}

type Config struct {
	URL    string
	Format string
	Path   string
	IPv6   bool
}

type Website struct {
	c *Config
}

func NewWebsite(c *Config) *Website {
	return &Website{c: c}
}

func (w *Website) Get(ctx context.Context) (utils.IP, error) {
	return w.roundtrip(ctx, -1, w.c.URL, w.c.Format, w.c.Path)
}

func (w *Website) roundtrip(ctx context.Context, i int, url, format, path string) (utils.IP, error) {
	ipr := utils.IP{}
	st := time.Now()
	log.Printf(`roundtrip: [%d] sending request to %s`, i, url)
	defer func() {
		et := time.Now()
		log.Printf(`roundtrip: [%d] time elapsed: %v`, i, et.Sub(st))
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ipr, fmt.Errorf(`roundtrip: bad request: %w`, err)
	}
	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				n := `tcp`
				if w.c.IPv6 {
					n = `tcp6`
				}
				return net.Dial(n, addr)
			},
		},
	}
	res, err := client.Do(req)
	if err != nil {
		return ipr, fmt.Errorf(`roundtrip: http err: %w`, err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		b, _ := io.ReadAll(io.LimitReader(res.Body, 1<<10))
		return ipr, fmt.Errorf(`roundtrip: bad status: %s: %s: %s`, res.Status, url, string(b))
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return ipr, fmt.Errorf(`roundtrip: error reading body: %w`, err)
	}
	str := string(body)

	var e Extractor
	switch format {
	case `json`:
		e = NewJsonExtractor(str, path)
	case `text`:
		e = NewRawExtractor(str)
	case `search`:
		e = NewSearchExtractor(str)
	default:
		return ipr, fmt.Errorf(`roundtrip: unknown type: %s`, format)
	}

	ipstr, err := e.Extract()
	if err != nil {
		return ipr, fmt.Errorf(`roundtrip: error extracting: %w`, err)
	}

	ip := net.ParseIP(ipstr)
	if len(ip) == net.IPv4len || ip.To4() != nil {
		ipr.V4 = ip.To4()
	} else if len(ip) == net.IPv6len {
		ipr.V6 = ip.To16()
	}

	if ipr.V4 == nil && ipr.V6 == nil {
		return ipr, fmt.Errorf(`no ipv4/ipv6 address was found`)
	}

	return ipr, nil
}
