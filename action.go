package filter

import (
	"regexp"
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
	wildcards  map[string]bool

	domainLists   ActionList
	hostsLists    ActionList
	regexLists    ActionList
	wildcardLists ActionList

	FileLoader FileListLoader
	HTTPLoader HTTPListLoader
}

// DNSNameRegexp matches valid domain names.
// Sourced from https://github.com/asaskevich/govalidator; const DNSName
//
// MIT Licensed, Copyright (c) 2014-2020 Alex Saskevich
var DNSNameRegexp = regexp.MustCompile(
	`^([a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62}){1}(\.[a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62})*[\._]?$`,
)

// HostsRegexp matches multiple spaces or tabspaces for cleaning up each entry
var HostsRegexp = regexp.MustCompile(`\s+|\t+`)

// NewActionConfig returns an action ready to accept configurations
func NewActionConfig(action ActionType) ActionConfig {
	return ActionConfig{
		configType:    action,
		domains:       make(map[string]bool),
		regex:         make(map[string]*regexp.Regexp),
		wildcards:     make(map[string]bool),
		domainLists:   make(ActionList),
		hostsLists:    make(ActionList),
		regexLists:    make(ActionList),
		wildcardLists: make(ActionList),
		FileLoader:    FileListLoader{},
		HTTPLoader:    HTTPListLoader{},
	}
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
	switch string(line) {
	case "[Adblock Plus]":
		// skip adblock plus file headers
		return true
	}
	return false
}
