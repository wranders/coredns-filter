package filter

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/coredns/caddy"
)

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

// AddHostsList to match contents
func (a ActionConfig) AddHostsList(url string) error {
	if _, ok := a.hostsLists[url]; !ok {
		loadFunc, err := a.GetListLoader(url)
		if err != nil {
			return err
		}
		a.hostsLists[url] = loadFunc
	}
	return nil
}

func (a ActionConfig) BuildHosts(domains map[string]bool) {
	for domain, loader := range a.hostsLists {
		file, err := loader.Load(domain)
		if err != nil {
			log.Errorf(
				"there was a problem fetching %s hosts list %q; %s",
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
			line = HostsRegexp.ReplaceAll(line, []byte(" "))
			hostsLine := strings.Split(string(line), " ")
			if len(hostsLine) != 2 {
				continue
			}
			domains[hostsLine[1]] = true
		}
	}
}
