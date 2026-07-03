package filter

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestLoadInvalidUrl(t *testing.T) {
	tests := []TestSetup{
		{
			"check load invalid url",
			`filter {
				allow list domain "https://not valid"
			}`,
			true,
		},
		{
			"check load missing host",
			`filter {
				allow list domain https://
			}`,
			true,
		},
	}
	for _, test := range tests {
		RunSetupTest(t, test)
	}
}

func TestLoadNonExistant(t *testing.T) {
	test := TestFilterBuild{
		"check non existant file",
		`filter {
				allow list domain file://noop
			}`,
		true,
	}
	RunFilterBuildTest(t, test)
}

func TestLoadAbsoluteFile(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("error getting working directory: %v", err)
	}

	test := TestFilterBuild{
		"check absolute file path",
		`filter {
				allow list domain file://` + path.Join(filepath.ToSlash(cwd), ".testdata/empty.list") + `
			}`,
		true,
	}
	RunFilterBuildTest(t, test)
}

func TestLoadNonExistantExternal(t *testing.T) {
	test := []TestFilterBuild{
		{
			"check non existant external resource",
			`filter {
				allow list domain http://noop
			}`,
			true,
		},
		{
			"check non existant external resource",
			`filter {
				allow list domain https://httpbin.org/status/404
			}`,
			true,
		},
	}
	for _, test := range test {
		RunFilterBuildTest(t, test)
	}
}

func TestListResolver(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping listresolver test; dns firewalling will cause this to fail")
	}
	http.DefaultTransport = nil
	corefile := `filter {
		listresolver 9.9.9.9
		block list domain https://s3.amazonaws.com/lists.disconnect.me/simple_tracking.txt
	}`
	filter := NewTestFilter(t, corefile)
	filter.Build()
	if len(filter.blockDomains) == 0 {
		t.Error("expected domains; got none")
	}
}

func TestListResolverTLS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping listresolver test; dns firewalling will cause this to fail")
	}
	http.DefaultTransport = nil
	corefile := `filter {
		listresolver tls://9.9.9.9 dns.quad9.net
		block list domain https://s3.amazonaws.com/lists.disconnect.me/simple_tracking.txt
	}`
	filter := NewTestFilter(t, corefile)
	filter.Build()
	if len(filter.blockDomains) == 0 {
		t.Errorf("expected domains; got none")
	}
}
