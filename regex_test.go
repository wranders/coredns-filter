package filter

import "testing"

func TestRegexNotProvided(t *testing.T) {
	test := TestSetup{
		"check regex not provided",
		`filter {
			block regex
		}`,
		true,
	}
	RunSetupTest(t, test)
}

func TestRegexInvalid(t *testing.T) {
	tests := []TestSetup{
		{
			"check regex invalid",
			`filter {
				allow regex *.example.net
			}`,
			true,
		},
		{
			"check regex invalid",
			`filter {
				block regex *.example.com
			}`,
			true,
		},
	}
	for _, test := range tests {
		RunSetupTest(t, test)
	}
}

func TestRegexAllowBlock(t *testing.T) {
	corefile := `filter {
		allow regex (^|[a-z0-9]+\.)sub.example.com
		block regex (^|.*\.)example.com
	}`
	tests := []TestFilterRequest{
		{
			"check blocked domain",
			"example.com",
			true,
		},
		{
			"check subdomains blocked",
			"some.example.com",
			true,
		},
		{
			"check allowed domain",
			"sub.example.com",
			false,
		},
		{
			"check implicity allowd safe subdomain",
			"safe.sub.example.com",
			false,
		},
	}
	RunFilterTests(t, corefile, tests)
}

func TestRegexListNotProvided(t *testing.T) {
	test := TestSetup{
		"check regex list not provided",
		`filter {
			block list regex
		}`,
		true,
	}
	RunSetupTest(t, test)
}

func TestRegexListInvalidScheme(t *testing.T) {
	tests := []TestSetup{
		{
			"check regex allow list invalid scheme",
			`filter {
				allow list regex scheme://noop
			}`,
			true,
		},
		{
			"check regex block list invalid scheme",
			`filter {
				block list regex scheme://noop
			}`,
			true,
		},
	}
	for _, test := range tests {
		RunSetupTest(t, test)
	}
}

func TestRegexListNonExistant(t *testing.T) {
	tests := []TestFilterBuild{
		{
			"check unreachable allow regex list",
			`filter {
				allow list regex https://noop
			}`,
			true,
		},
		{
			"check unreachable block regex list",
			`filter {
				block list regex https://noop
			}`,
			true,
		},
	}
	for _, test := range tests {
		RunFilterBuildTest(t, test)
	}
}

func TestRegexListAllowBlock(t *testing.T) {
	corefile := `filter {
		block list regex file://.testdata/regex.list
	}`
	tests := []TestFilterRequest{
		{
			"check allowed root domain",
			"example.com",
			false,
		},
		{
			"check subdomains blocked",
			"some.example.com",
			true,
		},
		{
			"check implicity allowd safe domain",
			"sub.example.net",
			false,
		},
	}
	RunFilterTests(t, corefile, tests)
}
