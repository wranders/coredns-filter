package filter

import (
	"testing"
)

func TestWildcardNotProvided(t *testing.T) {
	test := TestSetup{
		"check wildcard not provided",
		`filter {
			block wildcard
		}`,
		true,
	}
	RunSetupTest(t, test)
}

func TestWildcardBlockAllow(t *testing.T) {
	corefile := `filter {
		block wildcard example.com
		allow wildcard safe.example.com
	}`
	tests := []TestFilterRequest{
		{
			"check blocked domain",
			"example.com",
			true,
		},
		{
			"check blocked subdomain",
			"sub.example.com",
			true,
		},
		{
			"check explicity allowed domain",
			"safe.example.com",
			false,
		},
		{
			"check implicitly allowed subdomain",
			"sub.safe.example.com",
			false,
		},
	}
	RunFilterTests(t, corefile, tests)
}

func TestWildcardInvalid(t *testing.T) {
	tests := []TestSetup{
		{
			"check invalid block wildcard",
			`filter {
				block wildcard e[xample.com
			}`,
			true,
		},
		{
			"check invalid allow wildcard",
			`filter {
				allow wildcard e]xample.net
			}`,
			true,
		},
	}
	for _, test := range tests {
		RunSetupTest(t, test)
	}
}

func TestWildcardListExternal(t *testing.T) {
	// fetching the list from the source url works perfectly fine locally, but
	//   causes Github Actions to error stating:
	//   `http: no Client.Transport or DefaultTransport`, as if no RoundTripper
	//   exists in the default http.Client
	// this list is added to the included testdata until i can figure out what
	//   the hell is going on there...
	//
	// corefile := `filter {
	// 	block list wildcard https://small.oisd.nl
	// }`
	corefile := `filter {
		block list wildcard file://.testdata/oisd_small_abp.txt
	}`
	tests := []TestFilterRequest{
		{
			"check allowed domain",
			"youtube.com",
			false,
		},
		{
			"check blocked domain",
			"ads.youtube.com",
			true,
		},
	}
	RunFilterTests(t, corefile, tests)
}

func TestWildcardListBlockLocal(t *testing.T) {
	corefile := `filter {
		block list wildcard file://.testdata/wildcard.list
	}`
	tests := []TestFilterRequest{
		{
			"check first blocked domain",
			"example.com",
			true,
		},
		{
			"check second blocked domain",
			"example.net",
			true,
		},
		{
			"check non blocked domain",
			"example.org",
			false,
		},
	}
	RunFilterTests(t, corefile, tests)
}

func TestWildcardListAllowLocal(t *testing.T) {
	corefile := `filter {
		block wildcard example.org
		allow list wildcard file://.testdata/wildcard.list
	}`
	tests := []TestFilterRequest{
		{
			"check first allowed domain",
			"example.com",
			false,
		},
		{
			"check second allowed domain",
			"example.net",
			false,
		},
		{
			"check blocked domain",
			"example.org",
			true,
		},
	}
	RunFilterTests(t, corefile, tests)
}

func TestWildcardListExternalNotProvided(t *testing.T) {
	test := TestSetup{
		"check wildcard list source not provided",
		`filter {
			block list wildcard
		}`,
		true,
	}
	RunSetupTest(t, test)
}

func TestWildcardListNonExistant(t *testing.T) {
	test := TestFilterBuild{
		"check unreachable wildcard list",
		`filter {
			block list wildcard https://noop
		}`,
		true,
	}
	RunFilterBuildTest(t, test)
}

func TestWildcardListInvalidScheme(t *testing.T) {
	tests := []TestSetup{
		{
			"check allow invalid external wildcard list scheme",
			`filter {
				allow list wildcard scheme://noop
			}`,
			true,
		},
		{
			"check block invalid external wildcard list scheme",
			`filter {
				block list wildcard scheme://noop
			}`,
			true,
		},
	}
	for _, test := range tests {
		RunSetupTest(t, test)
	}
}
