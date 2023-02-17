package filter

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// ActionType represents the action taken upon domains, expressions, or lists
type ActionType int

const (
	// ActionTypeAllow represents a domain, expression, or list that will not
	// be filtered, regardless if the domain or expression is set to be blocked
	ActionTypeAllow ActionType = iota

	// ActionTypeBlock represents a domain, expression, or list that will be
	// filtered
	ActionTypeBlock
)

// String returns the action type
func (a ActionType) String() string {
	actions := map[ActionType]string{
		ActionTypeAllow: "allow",
		ActionTypeBlock: "block",
	}
	return actions[a]
}

// ActionList is list of URLs and the functions required to load them
type ActionList map[string]ListLoader

// ActionConfig contains the raw domains, expressions and lists that are
// compiled and used by the Filter
type ActionConfig struct {
	configType ActionType
	domains    map[string]bool
	regex      map[string]*regexp.Regexp

	domainLists   ActionList
	hostsLists    ActionList
	regexLists    ActionList
	wildcardLists ActionList

	FileLoader FileListLoader
	HTTPLoader HTTPListLoader

	hostsRegexp *regexp.Regexp
}

// DNSNameRegexp matches valid domain names.
// Sourced from https://github.com/asaskevich/govalidator; const DNSName
//
// MIT Licensed, Copyright (c) 2014-2020 Alex Saskevich
var DNSNameRegexp = regexp.MustCompile(
	`^([a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62}){1}(\.[a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62})*[\._]?$`,
)

// NewActionConfig returns an action ready to accept configurations
func NewActionConfig(action ActionType) ActionConfig {
	return ActionConfig{
		configType:    action,
		domains:       make(map[string]bool),
		regex:         make(map[string]*regexp.Regexp),
		domainLists:   make(ActionList),
		hostsLists:    make(ActionList),
		regexLists:    make(ActionList),
		wildcardLists: make(ActionList),
		FileLoader:    FileListLoader{},
		HTTPLoader:    HTTPListLoader{},
		hostsRegexp:   regexp.MustCompile(`\s+|\t+`),
	}
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

// AddRegex to match
func (a ActionConfig) AddRegex(expr string) error {
	comp, err := regexp.Compile(expr)
	if err != nil {
		return err
	}
	if _, ok := a.regex[expr]; !ok {
		a.regex[expr] = comp
	}
	return nil
}

// AddRegexList to match contents
func (a ActionConfig) AddRegexList(url string) error {
	if _, ok := a.regexLists[url]; !ok {
		loadFunc, err := a.GetListLoader(url)
		if err != nil {
			return err
		}
		a.regexLists[url] = loadFunc
	}
	return nil
}

// AddWildcard to match
func (a ActionConfig) AddWildcard(wildcard string) error {
	wc := a.makeWildcard(wildcard)
	if err := a.AddRegex(wc); err != nil {
		return err
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

func (ActionConfig) cleanListLine(line string) string {
	out := strings.TrimPrefix(line, ".")       // remove pcre-style wildcard prefix
	out = strings.TrimPrefix(out, "*.")        // remove generic wildcard
	out = strings.TrimPrefix(out, "||")        // remove adblock plus prefix
	out = strings.TrimSuffix(out, "^")         // remove adblock plus suffix
	out = strings.TrimPrefix(out, "address=/") // remove dnsmasq declaration
	out = strings.Split(out, "/")[0]           // remove dnsmasq ip address, if any
	return out
}

func (a ActionConfig) makeWildcard(expr string) string {
	out := strings.ReplaceAll(expr, ".", "\\.") // escape periods in domain name
	out = fmt.Sprintf("^.*\\.%s|^%s", out, out) // format to match root and sub domains
	return out
}

// BuildDomains creates a map of unique domains from explicit declarations and
// lists
func (a ActionConfig) BuildDomains() map[string]bool {
	domains := make(map[string]bool)

	for k := range a.domains {
		domains[k] = true
	}

	a.buildDomainsLists(domains)
	a.buildDomainHostsLists(domains)

	return domains
}

func (a ActionConfig) buildDomainsLists(domains map[string]bool) {
	for dom, loader := range a.domainLists {
		file, err := loader.Load(dom)
		if err != nil {
			log.Errorf(
				"there was a problem fetching %s domain list %q; %s",
				a.configType,
				dom,
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
				// Skip formats that contain whitespace characters
				// Zone files should use the stock 'file' plugin
				continue
			}
			domains[string(line)] = true
		}
	}
}

func (a ActionConfig) buildDomainHostsLists(domains map[string]bool) {
	for dom, loader := range a.hostsLists {
		file, err := loader.Load(dom)
		if err != nil {
			log.Errorf(
				"there was a problem fetching %s hosts list %q; %s",
				a.configType,
				dom,
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
			line = a.hostsRegexp.ReplaceAll(line, []byte(" "))
			hostsLine := strings.Split(string(line), " ")
			if len(hostsLine) != 2 {
				continue
			}
			domains[hostsLine[1]] = true
		}
	}
}

// BuildRegExps consolidates individual regular expressions then loads and
// compiles regular expressions from any configured lists
func (a ActionConfig) BuildRegExps() []*regexp.Regexp {

	regexes := make(map[string]*regexp.Regexp)
	for expr, regex := range a.regex {
		regexes[expr] = regex
	}

	a.buildRegExpsRegex(regexes)
	a.buildRegExpsWildcard(regexes)

	out := make([]*regexp.Regexp, 0, len(regexes))
	for _, expr := range regexes {
		out = append(out, expr)
	}

	return out
}

func (a ActionConfig) buildRegExpsRegex(r map[string]*regexp.Regexp) error {
	for dom, loader := range a.regexLists {
		file, err := loader.Load(dom)
		if err != nil {
			log.Errorf(
				"there was a problem fetching %s regex list %q; %s",
				a.configType,
				dom,
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
			lineStr := string(line)
			if _, ok := r[lineStr]; ok {
				continue
			}
			exp, err := regexp.Compile(lineStr)
			if err != nil {
				log.Errorf(
					"error compiling %s regular expression %q from list %s; %s",
					a.configType,
					lineStr,
					dom,
					err,
				)
			} else {
				r[lineStr] = exp
			}
		}
	}
	return nil
}

func (a ActionConfig) buildRegExpsWildcard(r map[string]*regexp.Regexp) error {
	for dom, loader := range a.wildcardLists {
		file, err := loader.Load(dom)
		if err != nil {
			log.Errorf(
				"there was a problem fetching %s domain list %q; %s",
				a.configType,
				dom,
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

			var ok bool
			var expString string
			if ok, expString = a.getWildcardString(line); !ok {
				continue
			}
			if _, ok := r[expString]; ok {
				continue
			}

			exp, err := regexp.Compile(expString)
			if err != nil {
				log.Errorf(
					"error compiling %s wildcard %q { %s } from list %s; %s",
					a.configType,
					a.cleanListLine(string(line)),
					expString,
					dom,
					err,
				)
				continue
			}
			r[expString] = exp
		}
	}
	return nil
}

func (a ActionConfig) getWildcardString(line []byte) (bool, string) {
	// strip comments and platform-specific formatting
	if a.shouldSkip(line) {
		return false, ""
	}
	clean := a.cleanListLine(string(line))
	if !DNSNameRegexp.MatchString(clean) {
		return false, ""
	}
	return true, a.makeWildcard(clean)
}

func (a ActionConfig) shouldSkip(line []byte) bool {
	if len(line) == 0 {
		// skip empty lines
		return true
	}
	switch line[0] {
	case '!', '#', ';':
		// skip commented lines
		return true
	}
	return false
}
