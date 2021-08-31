package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func request(ctx context.Context, sources []*SourceConfig, concurrency int) (string, error) {
	nSources := len(sources)
	if nSources <= 0 {
		return ``, fmt.Errorf(`no sources to request`)
	}
	if concurrency < 1 {
		panic(`invalid concurrency`)
	}
	if concurrency > nSources {
		concurrency = nSources
	}
	log.Printf(`concurrent requests: %d total: %d`, concurrency, nSources)

	type _Response struct {
		i   int
		ip  string
		err error
	}

	res := make(chan _Response)
	defer close(res)

	for next := unique(concurrency, nSources); ; {
		n := next()
		if n == -1 {
			break
		}
		s := sources[n]
		go func(i int, s *SourceConfig) {
			ip, err := roundtrip(ctx, i, s)
			if err != nil {
				log.Println(err)
			}
			res <- _Response{
				i:   i,
				ip:  ip,
				err: err,
			}
		}(n, s)
	}

	ips := make(map[string][]int)
	for i := 0; i < concurrency; i++ {
		r := <-res
		if r.err != nil {
			continue
		}
		ips[r.ip] = append(ips[r.ip], r.i)
	}

	ip, max := ``, 0
	for k, v := range ips {
		log.Printf(`IP: %-15s Count: %-d From: %v`, k, len(v), v)
		if len(v) > max {
			max = len(v)
			ip = k
		}
	}

	switch {
	case concurrency == 1 && max == 0:
		return ``, fmt.Errorf(`concurrency == 1 && success == 0`)
	case concurrency == 2 && max != 2:
		return ``, fmt.Errorf(`concurrency == 2 && not all succeeded`)
	default:
		if max < concurrency/2 {
			return ``, fmt.Errorf(`cannot find majority (%d/%d)`, max, concurrency)
		}
	}

	return ip, nil
}

// unique uniquely generates n integers within [0,N).
func unique(n int, N int) func() int {
	if n < 1 {
		return func() int {
			return -1
		}
	}
	chose := make(map[int]struct{}, n)
	return func() int {
		if len(chose) == n {
			return -1
		}
		var i int
		for {
			i = rand.Intn(N)
			if _, ok := chose[i]; ok {
				continue
			}
			chose[i] = struct{}{}
			break
		}
		return i
	}
}

func roundtrip(ctx context.Context, i int, s *SourceConfig) (string, error) {
	st := time.Now()
	log.Printf(`roundtrip: [%d] sending request to %s`, i, s.URL)
	defer func() {
		et := time.Now()
		log.Printf(`roundtrip: [%d] time elapsed: %v`, i, et.Sub(st))
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, nil)
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
		return ``, fmt.Errorf(`roundtrip: bad status: %s: %s: %s`, res.Status, s.URL, string(b))
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return ``, fmt.Errorf(`roundtrip: error reading body: %w`, err)
	}
	str := string(body)

	var e Extractor
	switch s.Type {
	case `json`:
		e = NewJsonExtractor(str, s.Path)
	case `raw`:
		e = NewRawExtractor(str)
	case `search`:
		e = NewSearchExtractor(str)
	default:
		return ``, fmt.Errorf(`roundtrip: unknown type: %s`, s.Type)
	}

	ip, err := e.Extract()
	if err != nil {
		return ``, fmt.Errorf(`roundtrip: error extracting: %w`, err)
	}

	return ip, nil
}
