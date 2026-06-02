package filter

import (
	"context"
	"net/netip"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

var log = clog.NewWithPlugin("filter")

// Filter checks if requested domains are blocked then returns the configured
// response
type Filter struct {
	Next plugin.Handler

	sync.RWMutex

	allowConfig    ActionConfig
	allowDomains   map[string]bool
	allowRegex     []*regexp.Regexp
	allowWildcards map[string]bool

	blockConfig    ActionConfig
	blockDomains   map[string]bool
	blockRegex     []*regexp.Regexp
	blockWildcards map[string]bool

	response Response

	startupOnce    sync.Once
	updateInterval time.Duration
	updateShutdown chan bool
}

func newFilter() *Filter {
	return &Filter{
		allowConfig:    NewActionConfig(ActionTypeAllow),
		allowDomains:   make(map[string]bool),
		allowRegex:     make([]*regexp.Regexp, 0),
		allowWildcards: make(map[string]bool),
		blockConfig:    NewActionConfig(ActionTypeBlock),
		blockDomains:   make(map[string]bool),
		blockRegex:     make([]*regexp.Regexp, 0),
		blockWildcards: make(map[string]bool),
		response: RespAddress{
			IP4: netip.IPv4Unspecified(),
			IP6: netip.IPv6Unspecified(),
		},
		updateInterval: 24 * time.Hour,
		updateShutdown: make(chan bool),
	}
}

// Name implements the plugin.Handler interface
// Returns the name of the CoreDNS plugin, in this case "filter"
func (f *Filter) Name() string {
	return "filter"
}

// ServeDNS implements the plugin.Handler inteface
// Checks whether or not the requested domain is allowed or blocked.
// Allowed domains are passed to the next plugin in the Corefile. Blocked
// domains return the configured response.
func (f *Filter) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	qname := strings.TrimSuffix(state.Name(), ".")

	var allowed, blocked bool
	f.RLock()
	allowed = f.isAllowed(qname)
	if !allowed {
		blocked = f.isBlocked(qname)
	}
	f.RUnlock()

	if !allowed && blocked {
		log.Debugf("blocking %q", qname)
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.RecursionAvailable = false
		response := f.response.Render(state.Name(), state.QType())
		msg.Authoritative = response.Authoritative
		msg.Answer = response.Answer
		w.WriteMsg(msg)
		return response.RCode, nil
	}

	return plugin.NextOrFailure(state.Name(), f.Next, ctx, w, r)
}

func (f *Filter) isAllowed(qname string) bool {
	if _, ok := f.allowDomains[qname]; ok {
		log.Debugf("request %q matched allowed domain", qname)
		return true
	}

	if wildcard, ok := matchesAnyWildcard(qname, f.allowWildcards); ok {
		log.Debugf("request %q matched allow wildcard %q", qname, wildcard)
		return true
	}

	// Evaluate regular expressions last, as they're the most expensive
	for _, exp := range f.allowRegex {
		if exp.MatchString(qname) {
			log.Debugf("request %q mached allow regex", qname)
			return true
		}
	}

	return false
}

func (f *Filter) isBlocked(qname string) bool {
	if _, ok := f.blockDomains[qname]; ok {
		log.Debugf("request %q matched blocked domain", qname)
		return true
	}

	if wildcard, ok := matchesAnyWildcard(qname, f.blockWildcards); ok {
		log.Debugf("request %q matched block wildcard %q", qname, wildcard)
		return true
	}

	// Evaluate regular expressions last, as they're the most expensive
	for _, exp := range f.blockRegex {
		if exp.MatchString(qname) {
			log.Debugf("request %q matched block regex", qname)
			return true
		}
	}

	return false
}

func matchesAnyWildcard(qname string, wildcards map[string]bool) (string, bool) {
	if wildcards[qname] {
		return qname, true
	}

	// Wildcards only match at subdomain boundaries. Test the qname with each
	// subdomain removed until we either match or run out of subdomains.
	for i, c := range qname {
		if c == '.' {
			wildcard := qname[i+1:]
			if wildcards[wildcard] {
				return wildcard, true
			}
		}
	}
	return "", false
}

// OnShutdown cleans up the filter and prepares it for removal
func (f *Filter) OnShutdown() error {
	if 0 < f.updateInterval {
		f.updateShutdown <- true
	}
	return nil
}

// Build the domain and regular expression lists used to determine how domains
// are handled
func (f *Filter) Build() {
	var allowDomains = make(map[string]bool)
	f.allowConfig.BuildDomains(allowDomains)
	f.allowConfig.BuildHosts(allowDomains)

	var blockDomains = make(map[string]bool)
	f.blockConfig.BuildDomains(blockDomains)
	f.blockConfig.BuildHosts(blockDomains)

	var allowRegexBuilder = make(map[string]*regexp.Regexp)
	f.allowConfig.BuildRegExps(allowRegexBuilder)
	allowRegex := f.consolidateRegex(allowRegexBuilder)

	var blockRegexBuilder = make(map[string]*regexp.Regexp)
	f.blockConfig.BuildRegExps(blockRegexBuilder)
	blockRegex := f.consolidateRegex(blockRegexBuilder)

	var allowWildcards = make(map[string]bool)
	f.allowConfig.BuildWildcards(allowWildcards)

	var blockWildcards = make(map[string]bool)
	f.blockConfig.BuildWildcards(blockWildcards)

	f.Lock()
	f.allowDomains = allowDomains
	f.allowRegex = allowRegex
	f.allowWildcards = allowWildcards
	f.blockDomains = blockDomains
	f.blockRegex = blockRegex
	f.blockWildcards = blockWildcards
	f.Unlock()

	log.Infof(
		"Successfully updated filter; "+
			"%d allowed domains, %d allowed regular expressions, %d allowed wildcards; "+
			"%d blocked domains, %d blocked regular expressions, %d blocked wildcards",
		len(f.allowDomains),
		len(f.allowRegex),
		len(f.allowWildcards),
		len(f.blockDomains),
		len(f.blockRegex),
		len(f.blockWildcards),
	)
}

func (f *Filter) consolidateRegex(regexes map[string]*regexp.Regexp) []*regexp.Regexp {
	out := make([]*regexp.Regexp, 0, len(regexes))
	for _, expr := range regexes {
		out = append(out, expr)
	}
	return out
}

// InitUpdate starts the update timer. This should only be run once on startup.
func (f *Filter) InitUpdate() error {
	if f.updateInterval == 0 {
		return nil
	}

	tick := time.NewTicker(f.updateInterval)

	go func() {
		for {
			select {
			case <-tick.C:
				f.Build()
			case <-f.updateShutdown:
				tick.Stop()
				return
			}
		}
	}()

	return nil
}
