package website

import (
	"context"
	"errors"
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
}

type Website struct {
	c *Config
}

func NewWebsite(c *Config) *Website {
	return &Website{c: c}
}

func (w *Website) Get(ctx context.Context) (utils.IP, error) {
	var je error
	ipr := utils.IP{}
	body1, err1 := w.roundtrip(ctx, -1, `tcp4`, w.c.URL)
	if err1 == nil {
		if ipstr, err := w.extract(body1, w.c.Format, w.c.Path); err1 == nil {
			if ip := net.ParseIP(ipstr); ip.To4() != nil {
				ipr.V4 = ip.To4()
			} else {
				log.Println(`invalid ipv4 address:`, ipstr, len(ip))
			}
		} else {
			je = errors.Join(je, err)
			log.Println(`error extracting from body:`, err, body1)
		}
	} else {
		log.Println(`error roundtripping for:`, err1, w.c.URL)
		je = errors.Join(je, err1)
	}
	body2, err2 := w.roundtrip(ctx, -1, `tcp6`, w.c.URL)
	if err2 == nil {
		if ipstr, err := w.extract(body2, w.c.Format, w.c.Path); err2 == nil {
			if ip := net.ParseIP(ipstr); ip.To16() != nil && ip.To4() == nil && ip.IsGlobalUnicast() {
				ipr.V6 = ip.To16()
			} else {
				log.Println(`invalid ipv6 address:`, ipstr)
			}
		} else {
			log.Println(`error extracting from body:`, err, body2)
			je = errors.Join(je, err)
		}
	} else {
		log.Println(`error roundtripping for:`, err2, w.c.URL)
		je = errors.Join(je, err2)
	}
	if ipr.V4 == nil && ipr.V6 == nil {
		return ipr, je
	}

	return ipr, nil
}

func (w *Website) roundtrip(ctx context.Context, i int, proto, url string) (string, error) {
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
	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial(proto, addr)
			},
		},
	}
	res, err := client.Do(req)
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

	return string(body), nil
}

func (w *Website) extract(body string, format, path string) (string, error) {
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
		return ``, fmt.Errorf(`roundtrip: unknown type: %s`, format)
	}

	ipstr, err := e.Extract()
	if err != nil {
		return ``, fmt.Errorf(`roundtrip: error extracting: %w`, err)
	}

	return ipstr, nil
}
