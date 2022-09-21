package filter

import (
	"context"
	"net/netip"
	"testing"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
)

func TestFilterResponseNoData(t *testing.T) {
	corefile := `filter {
		block domain example.com
		response nodata
	}`

	filter := NewTestFilter(t, corefile)
	filter.Build()
	req4 := new(dns.Msg).SetQuestion("example.com", dns.TypeA)
	rec4 := dnstest.NewRecorder(&test.ResponseWriter{})
	filter.ServeDNS(context.Background(), rec4, req4)
	if len(rec4.Msg.Answer) != 0 {
		t.Error(
			"error: ServeDNS error nodata4, response should contain no records",
		)
	}
	req6 := new(dns.Msg).SetQuestion("example.com", dns.TypeAAAA)
	rec6 := dnstest.NewRecorder(&test.ResponseWriter6{})
	filter.ServeDNS(context.Background(), rec6, req6)
	if len(rec4.Msg.Answer) != 0 {
		t.Error(
			"error: ServeDNS error nodata6, response should contain no records",
		)
	}
}

func TestFilterResponseNXDomain(t *testing.T) {
	corefile := `filter {
		block domain example.com
		response nxdomain
	}`

	filter := NewTestFilter(t, corefile)
	filter.Build()
	req4 := new(dns.Msg).SetQuestion("example.com", dns.TypeA)
	rec4 := dnstest.NewRecorder(&test.ResponseWriter{})
	filter.ServeDNS(context.Background(), rec4, req4)
	if len(rec4.Msg.Answer) == 0 {
		t.Error(
			"error: ServeDNS error nodata4, response should contain no records",
		)
	}
	req6 := new(dns.Msg).SetQuestion("example.com", dns.TypeAAAA)
	rec6 := dnstest.NewRecorder(&test.ResponseWriter6{})
	filter.ServeDNS(context.Background(), rec6, req6)
	if len(rec4.Msg.Answer) == 0 {
		t.Error(
			"error: ServeDNS error nodata6, response should contain no records",
		)
	}
}

func TestFilterResponseAddress(t *testing.T) {
	corefile := `filter {
		block domain example.com
		response address A 192.168.1.1 AAAA 2001:db8::0:1
	}`

	filter := NewTestFilter(t, corefile)
	filter.Build()

	req4 := new(dns.Msg).SetQuestion("example.com", dns.TypeA)
	rec4 := dnstest.NewRecorder(&test.ResponseWriter{})
	filter.ServeDNS(context.TODO(), rec4, req4)
	if len(rec4.Msg.Answer) == 0 {
		t.Error(
			"Error: ServeDNS error address, response should contain records",
		)
	}
	resp4 := netip.MustParseAddr(rec4.Msg.Answer[0].(*dns.A).A.String())
	expect4 := netip.MustParseAddr("192.168.1.1")
	if resp4 != expect4 {
		t.Error("Error: ServeDNS error address, A response should match config")
	}

	req6 := new(dns.Msg).SetQuestion("example.com", dns.TypeAAAA)
	rec6 := dnstest.NewRecorder(&test.ResponseWriter6{})
	filter.ServeDNS(context.TODO(), rec6, req6)
	if len(rec6.Msg.Answer) == 0 {
		t.Error(
			"Error: ServeDNS error nodata, response should contain records",
		)
	}
	resp6 := netip.MustParseAddr(rec6.Msg.Answer[0].(*dns.AAAA).AAAA.String())
	expect6 := netip.MustParseAddr("2001:db8::0:1")
	if resp6 != expect6 {
		t.Error("Error: ServeDNS error address, AAAA response should match config")
	}
}

func TestFilterResponseNonAddressQuestion(t *testing.T) {
	corefile := `filter {
		block domain example.com
	}`

	filter := NewTestFilter(t, corefile)
	filter.Build()
	req := new(dns.Msg).SetQuestion("example.com", dns.TypeCNAME)
	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	filter.ServeDNS(context.Background(), rec, req)
	if len(rec.Msg.Answer) != 0 {
		t.Error(
			"error: ServeDNS error nodata4, response should contain no records",
		)
	}
}
