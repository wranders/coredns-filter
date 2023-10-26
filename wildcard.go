package filter

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/coredns/caddy"
)

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

// AddWildcard to match
func (a ActionConfig) AddWildcard(wildcard string) error {
	wc := a.cleanWildcardListLine(wildcard)
	if !DNSNameRegexp.MatchString(wc) {
		errString := fmt.Sprintf(
			"wildcard %q is invalid",
			wildcard,
		)
		return errors.New(errString)
	}
	if _, ok := a.wildcards[wc]; !ok {
		a.wildcards[wc] = true
	}
	return nil
}

// AddWildcardList to match contents
func (a ActionConfig) AddWildcardList(url string) error {
	if _, ok := a.wildcardLists[url]; !ok {
		loadFunc, err := a.GetListLoader(url)
		if err != nil {
			return err
		}
		a.wildcardLists[url] = loadFunc
	}
	return nil
}

func (a ActionConfig) BuildWildcards(wildcards map[string]bool) {
	for wildcard := range a.wildcards {
		wildcards[wildcard] = true
	}

	for domain, loader := range a.wildcardLists {
		file, err := loader.Load(domain)
		if err != nil {
			log.Errorf(
				"there was a problem fetching %s wildcard list %q; %s",
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
			clean := a.cleanWildcardListLine(string(line))
			if !DNSNameRegexp.MatchString(clean) {
				log.Debugf(
					"wildcard %q is invalid",
					clean,
				)
				continue
			}
			if _, ok := wildcards[clean]; !ok {
				wildcards[clean] = true
			}
		}
	}
}

func (ActionConfig) cleanWildcardListLine(line string) string {
	out := strings.TrimPrefix(line, ".")       // remove pcre-style wildcard prefix
	out = strings.TrimPrefix(out, "*.")        // remove generic wildcard
	out = strings.TrimPrefix(out, "||")        // remove adblock plus prefix
	out = strings.TrimSuffix(out, "^")         // remove adblock plus suffix
	out = strings.TrimPrefix(out, "address=/") // remove dnsmasq declaration
	out = strings.Split(out, "/")[0]           // remove dnsmasq ip address, if any
	return out
}
