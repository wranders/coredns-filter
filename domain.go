package filter

import (
	"bufio"
	"bytes"

	"github.com/coredns/caddy"
)

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

// AddDomain to match
func (a ActionConfig) AddDomain(domain string) {
	if _, ok := a.domains[domain]; !ok {
		a.domains[domain] = true
	}
}

// AddDomainList to match contents
func (a ActionConfig) AddDomainList(url string) error {
	if _, ok := a.domainLists[url]; !ok {
		loadFunc, err := a.GetListLoader(url)
		if err != nil {
			return err
		}
		a.domainLists[url] = loadFunc
	}
	return nil
}

// BuildDomains creates a map of unique domains from explicit declarations and
// lists
func (a ActionConfig) BuildDomains(domains map[string]bool) {
	// populate single explicit domains
	for k := range a.domains {
		domains[k] = true
	}

	// populate domains from hosts files
	for domain, loader := range a.domainLists {
		file, err := loader.Load(domain)
		if err != nil {
			log.Errorf(
				"there was a problem fetching %s domain list: %q; %s",
				a.configType,
				domain,
				err,
			)
			continue
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			line := scanner.Bytes()
			line = bytes.TrimSpace(line)
			if a.shouldSkip(line) {
				continue
			}
			if bytes.Contains(line, []byte(" ")) {
				continue
			}
			domains[string(line)] = true
		}
	}
}
