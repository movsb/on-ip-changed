package getters

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"

	"github.com/movsb/on-ip-changed/utils"
)

func Request(ctx context.Context, getters []Getter, concurrency int) (utils.IP, error) {
	nSources := len(getters)
	if nSources <= 0 {
		return utils.IP{}, fmt.Errorf(`no sources to request`)
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
		ip  utils.IP
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

	ips4 := make(map[string][]int)
	ips6 := make(map[string][]int)
	for i := 0; i < concurrency; i++ {
		r := <-res
		if r.err != nil {
			continue
		}
		if r.ip.V4 != nil {
			ips4[r.ip.V4.String()] = append(ips4[r.ip.V4.String()], r.i)
		}
		if r.ip.V6 != nil {
			ips6[r.ip.V6.String()] = append(ips6[r.ip.V6.String()], r.i)
		}
	}

	ip4, max4 := ``, 0
	for k, v := range ips4 {
		log.Printf(`IP: %-39s Count: %-d From: %v`, k, len(v), v)
		if len(v) > max4 {
			max4 = len(v)
			ip4 = k
		}
	}
	ip6, max6 := ``, 0
	for k, v := range ips6 {
		log.Printf(`IP: %-39s Count: %-d From: %v`, k, len(v), v)
		if len(v) > max6 {
			max6 = len(v)
			ip6 = k
		}
	}

	switch {
	case concurrency == 1 && (max4 == 0 && max6 == 0):
		return utils.IP{}, fmt.Errorf(`concurrency == 1 && success == 0`)
	case concurrency == 2 && (max4 != 2 && max6 != 2):
		return utils.IP{}, fmt.Errorf(`concurrency == 2 && not all succeeded`)
	default:
		if max4 <= concurrency/2 && max6 <= concurrency/2 {
			return utils.IP{}, fmt.Errorf(`cannot find majority (4:%d/6:%d/%d)`, max4, max6, concurrency)
		}
	}

	return utils.IP{
		V4: net.ParseIP(ip4).To4(),
		V6: net.ParseIP(ip6).To16(),
	}, nil
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
