package website

import (
	"context"
	"testing"
)

func TestJsonExtractor(t *testing.T) {
	d := `{"a":{"rs":1,"code":0,"address":"中国  北京 北京市 教育网","ip":"103.201.26.28","isDomain":0}}`
	j := NewJsonExtractor(d, `a.ip`)
	ip, err := j.Extract()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip)
}

func TestRawExtractor(t *testing.T) {
	d := `   1.2.3.4  `
	j := NewRawExtractor(d)
	ip, err := j.Extract()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip)
}

func TestSearchExtractor(t *testing.T) {
	d := `  aaa 1.2.3.4  bbb `
	j := NewSearchExtractor(d)
	ip, err := j.Extract()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip)
}

func TestWebsites(t *testing.T) {
	w := NewWebsite(&Config{
		URL:    `https://ifconfig.co/ip`,
		Format: `text`,
	})
	t.Log(w.Get(context.TODO()))
}
