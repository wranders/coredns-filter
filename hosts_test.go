package filter

import "testing"

func TestHostsListNotProvided(t *testing.T) {
	test := TestSetup{
		"check hosts list not provided",
		`filter {
			block list hosts
		}`,
		true,
	}
	RunSetupTest(t, test)
}

func TestHostsListAllowBlock(t *testing.T) {
	corefile := `filter {
		block list hosts file://.testdata/hosts.list
		allow domain example.com
	}`
	tests := []TestFilterRequest{
		{
			"check domain blocked",
			"example.net",
			true,
		},
		{
			"check domain allowed",
			"example.com",
			false,
		},
	}
	RunFilterTests(t, corefile, tests)
}

func TestHostsListAllowPrecedence(t *testing.T) {
	corefile := `filter {
		allow list hosts file://.testdata/hosts.list
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

func TestHostsListNonExistant(t *testing.T) {
	tests := []TestFilterBuild{
		{
			"check allow invalid external hosts list scheme",
			`filter {
				allow list hosts https://noop
			}`,
			true,
		},
		{
			"check block invalid external hosts list scheme",
			`filter {
				block list hosts https://noop
			}`,
			true,
		},
	}
	for _, test := range tests {
		RunFilterBuildTest(t, test)
	}
}

func TestHostsListInvalidScheme(t *testing.T) {
	tests := []TestSetup{
		{
			"check allow invalid external hosts list scheme",
			`filter {
				allow list hosts scheme://noop
			}`,
			true,
		},
		{
			"check block invalid external hosts list scheme",
			`filter {
				block list hosts scheme://noop
			}`,
			true,
		},
	}
	for _, test := range tests {
		RunSetupTest(t, test)
	}
}
