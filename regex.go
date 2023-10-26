package filter

import (
	"bufio"
	"bytes"
	"regexp"

	"github.com/coredns/caddy"
)

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

// BuildRegExps consolidates individual regular expressions then loads and
// compiles regular expressions from any configured lists
func (a ActionConfig) BuildRegExps(regexps map[string]*regexp.Regexp) {
	for expression, regex := range a.regex {
		regexps[expression] = regex
	}

	for domain, loader := range a.regexLists {
		file, err := loader.Load(domain)
		if err != nil {
			log.Errorf(
				"there was a problem fetching %s regex list %q; %s",
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
			lineString := string(line)
			if _, ok := regexps[lineString]; ok {
				continue
			}
			expression, err := regexp.Compile(lineString)
			if err != nil {
				log.Debugf(
					"error compiling %s regular expression %q from list %q; %s",
					a.configType,
					lineString,
					domain,
					err,
				)
				continue
			}
			regexps[lineString] = expression
		}
	}
}
