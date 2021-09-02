package getters

import (
	"context"
	"fmt"
	"log"
	"math/rand"
)

func Request(ctx context.Context, getters []Getter, concurrency int) (string, error) {
	nSources := len(getters)
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
		s := getters[n]
		go func(i int, getter Getter) {
			ip, err := getter.Get(ctx)
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
		if max <= concurrency/2 {
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
