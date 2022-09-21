package filter

import (
	"net"
	"net/netip"

	"github.com/miekg/dns"
)

// RenderedResponse represents the information a response should return to the
// Filter.ServeDNS function
type RenderedResponse struct {
	// RCode is the DNS message response code
	RCode int

	// Authoritative signals if a response is authoritative.
	// Useful for NXDOMAIN responses
	Authoritative bool

	// Answer records if there are any
	Answer []dns.RR
}

// Response represents what ServeDNS expects to call
type Response interface {
	Render(qname string, qtype uint16) RenderedResponse
}

// RespAddress implements Response
// Return address records for IPv4 (A) and IPv6 (AAAA)
type RespAddress struct {
	IP4 netip.Addr
	IP6 netip.Addr
}

func (r RespAddress) Render(qname string, qtype uint16) RenderedResponse {
	var answer dns.RR
	header := dns.RR_Header{
		Name:   qname,
		Class:  dns.ClassINET,
		Ttl:    3600,
		Rrtype: qtype,
	}
	switch qtype {
	case dns.TypeA:
		answer = new(dns.A)
		answer.(*dns.A).Hdr = header
		answer.(*dns.A).A = net.IP(r.IP4.AsSlice())
	case dns.TypeAAAA:
		answer = new(dns.AAAA)
		answer.(*dns.AAAA).Hdr = header
		answer.(*dns.AAAA).AAAA = net.IP(r.IP6.AsSlice())
	default:
		return RenderedResponse{dns.RcodeSuccess, false, []dns.RR{}}
	}
	return RenderedResponse{dns.RcodeSuccess, false, []dns.RR{answer}}
}

// RespNoData implements Response
// Returns no records
type RespNoData struct{}

func (r RespNoData) Render(_ string, _ uint16) RenderedResponse {
	return RenderedResponse{dns.RcodeSuccess, false, []dns.RR{}}
}

// RespNXDomain implements Response
// Returns a dummy SOA record and an NXDOMAIN error code
type RespNXDomain struct{}

func (r RespNXDomain) Render(qname string, _ uint16) RenderedResponse {
	answer := &dns.SOA{
		Hdr: dns.RR_Header{
			Name:   qname,
			Rrtype: dns.TypeSOA,
			Class:  dns.ClassINET,
			Ttl:    0,
		},
		Ns:      "ns." + qname,
		Mbox:    "postmaster." + qname,
		Serial:  1,
		Refresh: 1,
		Retry:   1,
		Expire:  1,
		Minttl:  1,
	}
	return RenderedResponse{dns.RcodeNameError, true, []dns.RR{answer}}
}
