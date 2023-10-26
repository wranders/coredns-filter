package filter

import "testing"

func TestDomainNotProvided(t *testing.T) {
	test := TestSetup{
		"check domain not provided",
		`filter {
			block domain
		}`,
		true,
	}
	RunSetupTest(t, test)
}

func TestDomainAllowBlock(t *testing.T) {
	corefile := `filter {
		block wildcard example.com
		allow domain safe.example.com
		block domain unsafe.safe.example.com
	}`
	tests := []TestFilterRequest{
		{
			"check subdomains blocked",
			"sub.example.com",
			true,
		},
		{
			"check allowed domain",
			"safe.example.com",
			false,
		},
		{
			"check implicity blocked safe subdomain",
			"sub.safe.example.com",
			true,
		},
		{
			"check explicitly blocked safe subdomain",
			"unsafe.safe.example.com",
			true,
		},
	}
	RunFilterTests(t, corefile, tests)
}

func TestDomainListNotProvided(t *testing.T) {
	test := TestSetup{
		"check domain list not provided",
		`filter {
			allow list domain
		}`,
		true,
	}
	RunSetupTest(t, test)
}

func TestDomainListDeduplicated(t *testing.T) {
	corefile := `filter {
		block list domain file://.testdata/domain.list
	}`
	filter := NewTestFilter(t, corefile)
	filter.Build()
	numBlocked := len(filter.blockDomains)
	if numBlocked != 2 {
		t.Errorf(
			"error: domain: expected two (2) domains, got %d",
			numBlocked,
		)
	}
}

func TestDomainListAllowPrecedence(t *testing.T) {
	corefile := `filter {
		allow list domain file://.testdata/domain.list
		block domain example.com
	}`
	tests := []TestFilterRequest{
		{
			"check allow precedence",
			"example.com",
			false,
		},
		{
			"check allowed domain",
			"example.net",
			false,
		},
	}
	RunFilterTests(t, corefile, tests)
}

func TestDomainListNonExistant(t *testing.T) {
	test := TestFilterBuild{
		"check unreachable domain list",
		`filter {
			block list domain https://noop
		}`,
		true,
	}
	RunFilterBuildTest(t, test)
}

func TestDomainListInvalidScheme(t *testing.T) {
	tests := []TestSetup{
		{
			"check allow invalid external domain list scheme",
			`filter {
				allow list domain scheme://noop
			}`,
			true,
		},
		{
			"check block invalid external domain list scheme",
			`filter {
				block list domain scheme://noop
			}`,
			true,
		},
	}
	for _, test := range tests {
		RunSetupTest(t, test)
	}
}
