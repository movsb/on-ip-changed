package dnspod

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var dnsPodAPI = "https://dnsapi.cn"
var dnsPodUserAgentFormat = "ddns/0.0.0(%s)"

type Status struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (d *Status) Failed() bool {
	return d.Code != "1"
}

func (d *Status) Err() error {
	return errors.New(d.Message)
}

type Record struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ListRecordsResponse struct {
	Status  Status    `json:"status"`
	Records []*Record `json:"records"`
}

type CreateRecordsResponse struct {
	Status Status `json:"status"`
	Record Record `json:"record"`
}

type ModifyRecordsResponse struct {
	Status Status `json:"status"`
}

type Config struct {
	Token string `yaml:"token"`
	Email string `yaml:"email"`
}

type DnsPod struct {
	c *Config
}

func NewDnsPod(c *Config) *DnsPod {
	return &DnsPod{
		c: c,
	}
}

// CreateRecord creates a new record.
func (d *DnsPod) CreateRecord(ctx context.Context, domain string, record string, ty string, value string) (*Record, error) {
	data, err := d.post(ctx, `Record.Create`, map[string]interface{}{
		`domain`:      domain,
		`sub_domain`:  record,
		"record_type": ty,
		"record_line": "默认",
		`value`:       value,
	})
	if err != nil {
		return nil, fmt.Errorf("dnspod: error creating record: %w", err)
	}
	var resp CreateRecordsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("dnspod: error unmarshaling record: %w", err)
	}
	if resp.Status.Failed() {
		return nil, fmt.Errorf(`dnspod: %w`, resp.Status.Err())
	}
	return &resp.Record, nil
}

func (d *DnsPod) FindRecord(ctx context.Context, domain string, ty string, record string) (*Record, error) {
	data, err := d.post(ctx, "Record.List", map[string]interface{}{
		"domain":      domain,
		"sub_domain":  record,
		"record_type": ty,
	})
	if err != nil {
		return nil, fmt.Errorf("dnspod: error finding record: %w", err)
	}
	var resp ListRecordsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("dnspod: error unmarshaling record: %w", err)
	}
	if resp.Status.Failed() {
		return nil, fmt.Errorf("dnspod: error listing records: %w", resp.Status.Err())
	}
	if len(resp.Records) != 1 {
		return nil, fmt.Errorf("dnspod: record not found")
	}
	if resp.Records[0].Name != record {
		return nil, fmt.Errorf("dnspod: response record not match")
	}
	return resp.Records[0], nil
}

func (d *DnsPod) ModifyRecord(ctx context.Context, domain string, recordID string, record string, ty string, value string) error {
	data, err := d.post(ctx, "Record.Modify", map[string]interface{}{
		"domain":      domain,
		"record_id":   recordID,
		"sub_domain":  record,
		"record_type": ty,
		"record_line": "默认",
		"value":       value,
	})
	if err != nil {
		return fmt.Errorf(`dnspod: error modifying record: %w`, err)
	}
	var resp ModifyRecordsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf(`dnspod: error unmarshaling record: %w`, err)
	}
	if resp.Status.Failed() {
		return fmt.Errorf(`dnspod: status failed: %w`, resp.Status.Err())
	}
	return nil
}

func (d *DnsPod) post(ctx context.Context, method string, values map[string]interface{}) ([]byte, error) {
	v := url.Values{}
	v.Set("login_token", d.c.Token)
	v.Set("format", "json")
	for key, val := range values {
		v.Set(key, fmt.Sprint(val))
	}
	uri := fmt.Sprintf("%s/%s", dnsPodAPI, method)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", fmt.Sprintf(dnsPodUserAgentFormat, d.c.Email))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
