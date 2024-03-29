package asus

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/movsb/on-ip-changed/utils"
	"github.com/movsb/on-ip-changed/utils/registry"
)

func init() {
	registry.RegisterGetter(`asus`, Config{}, NewAsus)
}

type Config struct {
	Address  string `yaml:"address"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Asus struct {
	c     *Config
	token string
}

func NewAsus(c *Config) *Asus {
	return &Asus{c: c}
}

// IPv4 only
func (a *Asus) Get(ctx context.Context) (utils.IP, error) {
	if err := a.login(ctx); err != nil {
		return utils.IP{}, err
	}
	s, err := a.status(ctx)
	if err != nil {
		return utils.IP{}, err
	}
	return utils.IP{
		V4: net.ParseIP(s.WanLinkIpAddr).To4(),
	}, nil
}

func (a *Asus) login(ctx context.Context) error {
	u, err := url.Parse(utils.AddHTTPPrefix(a.c.Address))
	if err != nil {
		return fmt.Errorf(`asus: bad address: %s: %w`, a.c.Address, err)
	}
	u.Path = filepath.Join(`/`, u.Path, `login.cgi`)
	body := `group_id=&action_mode=&action_script=&action_wait=5&current_page=Main_Login.asp&next_page=index.asp&login_authorization=%s`
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`%s:%s`, a.c.Username, a.c.Password)))
	body = fmt.Sprintf(body, auth)
	log.Printf(`asus: url: %s`, u.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), strings.NewReader(body))
	if err != nil {
		return fmt.Errorf(`asus: bad request: %w`, err)
	}
	req.Header.Set(`Content-Type`, `application/x-www-form-urlencoded`)
	req.Header.Set(`Referer`, u.String()) // must
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf(`asus: http error: %w`, err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf(`asus: bad status: %s`, res.Status)
	}
	cookieName := `asus_token`
	cookieValue := ``
	for _, cookie := range res.Cookies() {
		if cookie.Name == cookieName {
			cookieValue = cookie.Value
			break
		}
	}
	if cookieValue == "" {
		body, _ := ioutil.ReadAll(res.Body)
		return fmt.Errorf(`asus: no cookie, response: %q`, string(body))
	}
	a.token = cookieValue
	return nil
}

type Status struct {
	WanLinkIpAddr string
}

func (a *Asus) status(ctx context.Context) (*Status, error) {
	u, err := url.Parse(utils.AddHTTPPrefix(a.c.Address))
	if err != nil {
		return nil, fmt.Errorf(`asus: bad address: %s: %w`, a.c.Address, err)
	}
	u.Path = filepath.Join(`/`, u.Path, `status.asp`)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf(`asus: bad request: %w`, err)
	}
	req.AddCookie(&http.Cookie{
		Name:  `asus_token`,
		Value: a.token,
	})
	req.Header.Set(`Referer`, u.String()) // must
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(`asus: http error: %w`, err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf(`asus: bad status: %s`, res.Status)
	}
	// function wanlink_ipaddr() { return '14.155.112.114';}
	re := regexp.MustCompile(`function wanlink_ipaddr\(\) \{ return '([^']+)';}`)
	body, _ := ioutil.ReadAll(res.Body)
	matches := re.FindStringSubmatch(string(body))
	if matches == nil || len(matches[1]) <= 0 {
		return nil, fmt.Errorf(`asus: no wan ip was found`)
	}
	ip := net.ParseIP(matches[1])
	if ip == nil || ip.To4() == nil || ip.IsUnspecified() {
		return nil, fmt.Errorf(`asus: no wan ip was found`)
	}
	s := &Status{
		WanLinkIpAddr: matches[1],
	}
	return s, nil
}
