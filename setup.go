package filter

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/parse"
	"github.com/miekg/dns"
)

func init() {
	plugin.Register("filter", setup)
}

func setup(c *caddy.Controller) error {
	f := newFilter()

	if err := Parse(c, f); err != nil {
		return err
	}

	c.OnShutdown(f.OnShutdown)
	c.OnStartup(func() error {
		f.startupOnce.Do(func() {
			f.Build()
			f.InitUpdate()
		})
		return nil
	})

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		f.Next = next
		return f
	})

	return nil
}

func ensureEOL(c *caddy.Controller) error {
	if remain := c.RemainingArgs(); len(remain) != 0 {
		return errorExpectedEOL{data: remain}
	}
	return nil
}

// Parse the Corefile configuration
func Parse(c *caddy.Controller, f *Filter) error {
	c.Next()
	for c.NextBlock() {
		switch c.Val() {
		case "allow":
			if err := parseAction(c, f, ActionTypeAllow); err != nil {
				return err
			}
		case "block":
			if err := parseAction(c, f, ActionTypeBlock); err != nil {
				return err
			}
		case "listresolver":
			if err := parseListResolver(c, f); err != nil {
				return err
			}
		case "response":
			if err := parseResponse(c, f); err != nil {
				return err
			}
		case "update":
			if !c.NextArg() {
				return c.Err("no update interval specified")
			}
			duration, err := time.ParseDuration(c.Val())
			if err != nil {
				return c.Errf("invalid update interval %q; %w", c.Val(), err)
			}
			f.updateInterval = duration
		default:
			return c.Errf(
				"unknown token %q;"+
					"expected 'allow', 'block', 'response', or 'update'",
				c.Val(),
			)
		}
	}
	return nil
}

func parseAction(c *caddy.Controller, f *Filter, a ActionType) error {
	if !c.NextArg() {
		return c.Errf(
			"no %s type specified; "+
				"expected 'domain', 'regex', 'wildcard', or 'list'",
			a,
		)
	}
	switch c.Val() {
	case "domain":
		if err := parseActionDomain(c, f, a); err != nil {
			return err
		}
	case "regex":
		if err := parseActionRegex(c, f, a); err != nil {
			return err
		}
	case "wildcard":
		if err := parseActionWildcard(c, f, a); err != nil {
			return err
		}
	case "list":
		if err := parseActionList(c, f, a); err != nil {
			return err
		}
	default:
		return c.Errf("unexpected %s token %q", a, c.Val())
	}
	return nil
}

func parseActionDomain(c *caddy.Controller, f *Filter, a ActionType) error {
	if !c.NextArg() {
		return c.Errf("no %s domain specified", a)
	}
	switch a {
	case ActionTypeAllow:
		f.allowConfig.AddDomain(c.Val())
	case ActionTypeBlock:
		f.blockConfig.AddDomain(c.Val())
	}
	return ensureEOL(c)
}

func parseActionRegex(c *caddy.Controller, f *Filter, a ActionType) error {
	if !c.NextArg() {
		return c.Errf("no %s regex specified", a)
	}
	switch a {
	case ActionTypeAllow:
		if err := f.allowConfig.AddRegex(c.Val()); err != nil {
			return err
		}
	case ActionTypeBlock:
		if err := f.blockConfig.AddRegex(c.Val()); err != nil {
			return err
		}
	}
	return ensureEOL(c)
}

func parseActionWildcard(c *caddy.Controller, f *Filter, a ActionType) error {
	if !c.NextArg() {
		return c.Errf("no %s regex specified", a)
	}
	switch a {
	case ActionTypeAllow:
		if err := f.allowConfig.AddWildcard(c.Val()); err != nil {
			return err
		}
	case ActionTypeBlock:
		if err := f.blockConfig.AddWildcard(c.Val()); err != nil {
			return err
		}
	}
	return ensureEOL(c)
}

func parseActionList(c *caddy.Controller, f *Filter, a ActionType) error {
	if !c.NextArg() {
		return c.Errf("no %s list type specified", a)
	}
	switch c.Val() {
	case "domain":
		if err := parseActionListDomain(c, f, a); err != nil {
			return err
		}
	case "hosts":
		if err := parseActionListHosts(c, f, a); err != nil {
			return err
		}
	case "regex":
		if err := parseActionListRegex(c, f, a); err != nil {
			return err
		}
	case "wildcard":
		if err := parseActionListWildcard(c, f, a); err != nil {
			return err
		}
	default:
		return c.Errf(
			"unexpected %s token %q; "+
				"expected 'domain', 'regex', or 'wildcard'",
			a,
			c.Val(),
		)
	}
	return nil
}

func parseActionListDomain(c *caddy.Controller, f *Filter, a ActionType) error {
	if !c.NextArg() {
		return c.Errf("no %s domain list specified", a)
	}
	switch a {
	case ActionTypeAllow:
		if err := f.allowConfig.AddDomainList(c.Val()); err != nil {
			return err
		}
	case ActionTypeBlock:
		if err := f.blockConfig.AddDomainList(c.Val()); err != nil {
			return err
		}
	}
	return ensureEOL(c)
}

func parseActionListHosts(c *caddy.Controller, f *Filter, a ActionType) error {
	if !c.NextArg() {
		return c.Errf("no %s hosts list specified", a)
	}
	switch a {
	case ActionTypeAllow:
		if err := f.allowConfig.AddHostsList(c.Val()); err != nil {
			return err
		}
	case ActionTypeBlock:
		if err := f.blockConfig.AddHostsList(c.Val()); err != nil {
			return err
		}
	}
	return ensureEOL(c)
}

func parseActionListRegex(c *caddy.Controller, f *Filter, a ActionType) error {
	if !c.NextArg() {
		return c.Errf("no %s regex list specified", a)
	}
	switch a {
	case ActionTypeAllow:
		if err := f.allowConfig.AddRegexList(c.Val()); err != nil {
			return err
		}
	case ActionTypeBlock:
		if err := f.blockConfig.AddRegexList(c.Val()); err != nil {
			return err
		}
	}
	return ensureEOL(c)
}

func parseActionListWildcard(c *caddy.Controller, f *Filter, a ActionType) error {
	if !c.NextArg() {
		return c.Errf("no %s wildcard list specified", a)
	}
	switch a {
	case ActionTypeAllow:
		if err := f.allowConfig.AddWildcardList(c.Val()); err != nil {
			return err
		}
	case ActionTypeBlock:
		if err := f.blockConfig.AddWildcardList(c.Val()); err != nil {
			return err
		}
	}
	return ensureEOL(c)
}

func parseResponse(c *caddy.Controller, f *Filter) error {
	if !c.NextArg() {
		return c.Err(
			"no response type specified; " +
				"expected 'address', 'nxdomain', 'nodata', or 'null'",
		)
	}
	r := strings.ToLower(c.Val())
	switch r {
	case "address":
		if err := parseResponseAddress(c, f); err != nil {
			return err
		}
	case "nxdomain":
		f.response = RespNXDomain{}
	case "nodata":
		f.response = RespNoData{}
	case "null":
		f.response = RespAddress{
			IP4: netip.IPv4Unspecified(),
			IP6: netip.IPv6Unspecified(),
		}
	default:
		return c.Errf(
			"unknown response type %q; "+
				"expected 'address', 'nxdomain', 'nodata', or 'null'",
			r,
		)
	}
	return nil
}

func parseResponseAddress(c *caddy.Controller, f *Filter) error {
	if !c.NextArg() {
		return c.Errf("no address records specified")
	}
	cur := []string{c.Val()}
	remaining := append(cur, c.RemainingArgs()...)
	if len(remaining) != 2 && len(remaining) != 4 {
		return c.Errf(
			"unexpected number of address arguments %q; "+
				"expected one of 'a' and/or 'aaaa' with associated addresses",
			remaining,
		)
	}

	// ip4 := netip.IPv4Unspecified()
	// ip6 := netip.IPv6Unspecified()
	resp := RespAddress{
		IP4: netip.IPv4Unspecified(),
		IP6: netip.IPv6Unspecified(),
	}

	firstRec, firstAddr, err := parseAddress(remaining[:2]...)
	if err != nil {
		return err
	}
	switch firstRec {
	case dns.TypeA:
		resp.IP4 = firstAddr
	case dns.TypeAAAA:
		resp.IP6 = firstAddr
	}

	if len(remaining) != 4 {
		f.response = resp
		return nil
	}

	secondRec, secondAddr, err := parseAddress(remaining[2:]...)
	if err != nil {
		return err
	}
	if firstRec == secondRec {
		return errors.New(
			"duplicate response address record type provided; " +
				"expected only one of each 'a' and 'aaaa'",
		)
	}
	switch secondRec {
	case dns.TypeA:
		resp.IP4 = secondAddr
	case dns.TypeAAAA:
		resp.IP6 = secondAddr
	}

	f.response = resp

	return nil
}

// parseAddress expects the rec parameter to be
//
//	[0]: record type (A or AAAA)
//	[1]: record address
func parseAddress(rec ...string) (uint16, netip.Addr, error) {
	recTypeLower := strings.ToLower(rec[0])
	addr, err := netip.ParseAddr(rec[1])
	if err != nil {
		return 0,
			netip.Addr{},
			fmt.Errorf(
				"invalid response %q record address %q; %w",
				rec[0],
				rec[1],
				err,
			)
	}
	if addr.Is4() {
		if strings.Compare(recTypeLower, "aaaa") == 0 {
			return 0,
				addr,
				fmt.Errorf(
					"response record:address type mismatch; "+
						"record is 'aaaa' with ipv4 %q",
					addr.String(),
				)
		}
		return dns.TypeA, addr, nil
	}
	if addr.Is6() {
		if strings.Compare(recTypeLower, "a") == 0 {
			return 0,
				addr,
				fmt.Errorf(
					"response record:address type mismatch; "+
						"record is 'a' with ipv6 %q",
					addr.String(),
				)
		}
		return dns.TypeAAAA, addr, nil
	}
	// This shouldn't be reached but it's here just in case something
	// catastrophic happens
	return 0, netip.Addr{}, fmt.Errorf("unknown response address %q", rec[1])
}

func parseListResolver(c *caddy.Controller, f *Filter) error {
	to := c.RemainingArgs()
	if len(to) == 0 {
		return c.Errf("no list resolver address specified")
	}
	toHosts, err := parse.HostPortOrFile(to...)
	if err != nil {
		return err
	}
	transport, host := parse.Transport(toHosts[0])
	ipaddr, err := netip.ParseAddrPort(host)
	if err != nil {
		return err
	}
	switch transport {
	case "dns":
		f.allowConfig.HTTPLoader.Network = "udp"
		f.blockConfig.HTTPLoader.Network = "udp"
	case "tls":
		f.allowConfig.HTTPLoader.Network = "tcp"
		f.blockConfig.HTTPLoader.Network = "tcp"
	default:
		return fmt.Errorf(
			"%q is not a supported transport for listresolver",
			transport,
		)
	}
	f.allowConfig.HTTPLoader.ResolverIP = ipaddr
	f.blockConfig.HTTPLoader.ResolverIP = ipaddr
	return nil
}
