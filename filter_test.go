package filter

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
)

type testFilterRequest struct {
	QName     string
	WantBlock bool
}

func NewTestFilter(t *testing.T, corefile string) *Filter {
	controller := caddy.NewTestController("dns", corefile)
	filter := newFilter()
	if err := Parse(controller, filter); err != nil {
		t.Error(err)
	}
	filter.Next = test.ErrorHandler()
	return filter
}

func TestFilterGetName(t *testing.T) {
	corefile := `filter`
	filter := NewTestFilter(t, corefile)
	if strings.Compare(filter.Name(), "filter") != 0 {
		t.Error("incorrect plugin name")
	}
}

func TestFilterResolve(t *testing.T) {
	corefile := `filter {
		block domain example.com
		allow domain one.example.com

		allow wildcard two.exampletwo.com
		block wildcard exampletwo.com
	}`
	tests := []testFilterRequest{
		{
			"example.com",
			true,
		},
		{
			"one.example.com",
			false,
		},
		{
			"www.example.com",
			false,
		},
		{
			"aexample.com",
			false,
		},
		{
			"www.exampletwo.com",
			true,
		},
		{
			"one.exampletwo.com",
			true,
		},
		{
			"sub.two.exampletwo.com",
			false,
		},
	}

	filter := NewTestFilter(t, corefile)
	filter.Build()
	for _, tt := range tests {
		t.Run(tt.QName, func(t *testing.T) {
			req4 := new(dns.Msg).SetQuestion(tt.QName, dns.TypeA)
			rec4 := dnstest.NewRecorder(&test.ResponseWriter{})
			filter.ServeDNS(context.Background(), rec4, req4)
			if (len(rec4.Msg.Answer) == 0) == tt.WantBlock {
				t.Errorf(
					"error: ServeDNS error = %s A, wantBlock = %v",
					tt.QName,
					tt.WantBlock,
				)
			}

			req6 := new(dns.Msg).SetQuestion(tt.QName, dns.TypeAAAA)
			rec6 := dnstest.NewRecorder(&test.ResponseWriter6{})
			filter.ServeDNS(context.Background(), rec6, req6)
			if (len(rec6.Msg.Answer) == 0) == tt.WantBlock {
				t.Errorf(
					"error: ServeDNS error = %s AAAA, wantBlock = %v",
					tt.QName,
					tt.WantBlock,
				)
			}
		})
	}
}

func TestFilterReload(t *testing.T) {
	corefile := `filter`

	filter := NewTestFilter(t, corefile)
	filter.updateInterval = 10 * time.Millisecond
	if err := filter.InitUpdate(); err != nil {
		t.Error(err)
	}
	time.Sleep(20 * time.Millisecond)
	filter.OnShutdown()

	filter.updateInterval = 0
	if err := filter.InitUpdate(); err != nil {
		t.Error(err)
	}
}
