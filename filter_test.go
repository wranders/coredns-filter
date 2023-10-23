package filter

import (
	"bytes"
	"context"
	glog "log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
)

type TestFilterBuild struct {
	Name     string
	Corefile string
	WantErr  bool
}

type TestFilterRequest struct {
	Name      string
	QName     string
	WantBlock bool
}

func RunFilterBuildTest(t *testing.T, testbuild TestFilterBuild) {
	controller := caddy.NewTestController("dns", testbuild.Corefile)
	filter := newFilter()
	if err := Parse(controller, filter); err != nil {
		t.Error(err)
	}
	filter.Next = test.ErrorHandler()
	t.Run(testbuild.Name, func(t *testing.T) {
		var buf bytes.Buffer
		glog.SetOutput(&buf)
		defer func() {
			glog.SetOutput(os.Stderr)
		}()
		filter.Build()
		if strings.Contains(buf.String(), "[ERROR]") && testbuild.WantErr == false {
			t.Errorf(
				"error: filter build: %s, wanterr: %t\ncorefile: %s",
				testbuild.Name,
				testbuild.WantErr,
				testbuild.Corefile,
			)
		}
	})
}

func RunFilterTests(t *testing.T, corefile string, tests []TestFilterRequest) {
	controller := caddy.NewTestController("dns", corefile)
	filter := newFilter()
	if err := Parse(controller, filter); err != nil {
		t.Error(err)
	}
	filter.Next = test.ErrorHandler()
	filter.Build()
	for _, tt := range tests {
		t.Run(tt.QName, func(t *testing.T) {
			req4 := new(dns.Msg).SetQuestion(tt.QName, dns.TypeA)
			rec4 := dnstest.NewRecorder(&test.ResponseWriter{})
			filter.ServeDNS(context.Background(), rec4, req4)
			if (len(rec4.Msg.Answer) == 0) == tt.WantBlock {
				t.Errorf(
					"error: %s A, wantblock: %t",
					tt.QName,
					tt.WantBlock,
				)
			}
			req6 := new(dns.Msg).SetQuestion(tt.QName, dns.TypeAAAA)
			rec6 := dnstest.NewRecorder(&test.ResponseWriter{})
			filter.ServeDNS(context.Background(), rec6, req6)
			if (len(rec4.Msg.Answer) == 0) == tt.WantBlock {
				t.Errorf(
					"error: %s AAAA, wantblock: %t",
					tt.QName,
					tt.WantBlock,
				)
			}
		})
	}
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
	// this exists purely for coverage
	corefile := `filter`
	filter := NewTestFilter(t, corefile)
	if strings.Compare(filter.Name(), "filter") != 0 {
		t.Error("incorrect plugin name")
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
