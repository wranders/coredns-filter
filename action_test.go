package filter

import (
	"net/http"
	"testing"
)

func TestActionList404(t *testing.T) {
	corefile := `filter {
		block list domain https://httpbin.org/status/404
		block list regex https://httpbin.org/status/404
		block list wildcard https://httpbin.org/noop
		block list hosts https://httpbin.org/status/404
	}`
	filter := NewTestFilter(t, corefile)
	filter.Build()
	// if filter.blockTree.Len() != 0 {
	// 	t.Error("block tree should be empty")
	// }
	if len(filter.blockDomains) != 0 {
		t.Error("block domains should be empty")
	}
	if len(filter.blockRegex) != 0 {
		t.Error("block regex should be empty")
	}
}

func TestActionListDomain(t *testing.T) {
	corefile := `filter {
		block list domain file://.testdata/domain.list
	}`
	filter := NewTestFilter(t, corefile)
	filter.Build()
	// if filter.blockTree.Len() != 2 {
	// 	t.Errorf(
	// 		"expected two domains; found %d",
	// 		filter.blockTree.Len(),
	// 	)
	// }
	if len(filter.blockDomains) != 2 {
		t.Errorf(
			"expected two domains; found %d",
			len(filter.blockDomains),
		)
	}
}

func TestActionListHosts(t *testing.T) {
	corefile := `filter {
		block list hosts file://.testdata/hosts.list
	}`
	filter := NewTestFilter(t, corefile)
	filter.Build()
	if len(filter.blockDomains) != 2 {
		t.Errorf(
			"expected two domains; found %d",
			len(filter.blockDomains),
		)
	}
}

func TestActionListRegex(t *testing.T) {
	corefile := `filter {
		block list regex file://.testdata/regex.list
	}`
	filter := NewTestFilter(t, corefile)
	filter.Build()
	if len(filter.blockRegex) != 1 {
		t.Errorf(
			"expected only one valid regexp; found %d",
			len(filter.blockRegex),
		)
	}
}

func TestActionListWildcard(t *testing.T) {
	corefile := `filter {
		block list wildcard file://.testdata/wildcard.list
	}`
	filter := NewTestFilter(t, corefile)
	filter.Build()
	if len(filter.blockRegex) != 2 {
		t.Errorf(
			"expected two valid regexp; found %d; %v",
			len(filter.blockRegex),
			filter.blockRegex,
		)
	}
}

func TestActionMissingFile(t *testing.T) {
	corefile := `filter {
		block list domain file://.testdata/noop.list
	}`
	filter := NewTestFilter(t, corefile)
	filter.Build()
	// if filter.blockTree.Len() != 0 {
	// 	t.Errorf(
	// 		"expected no domains; found %d",
	// 		filter.blockTree.Len(),
	// 	)
	// }
	if len(filter.blockDomains) != 0 {
		t.Errorf(
			"expected no domains; found %d",
			len(filter.blockDomains),
		)
	}
}

func TestActionHTTPList(t *testing.T) {
	corefile := `filter {
		block list domain https://dbl.oisd.nl/basic/
	}`
	filter := NewTestFilter(t, corefile)
	filter.Build()
	// if filter.blockTree.Len() == 0 {
	// 	t.Errorf(
	// 		"expected domains from remote list; found %d",
	// 		filter.blockTree.Len(),
	// 	)
	// }
	if len(filter.blockDomains) == 0 {
		t.Errorf(
			"expected domains from remote list; found %d",
			len(filter.blockDomains),
		)
	}
}

func TestActionNonExistentHTTPList(t *testing.T) {
	corefile := `filter {
		block list domain http://localhost/noop
	}`
	filter := NewTestFilter(t, corefile)
	filter.Build()
	// if filter.blockTree.Len() != 0 {
	// 	t.Errorf(
	// 		"expected no domains from non-existent remote list; found %d",
	// 		filter.blockTree.Len(),
	// 	)
	// }
	if len(filter.blockDomains) != 0 {
		t.Errorf(
			"expected no domains from non-existent remote list; found %d",
			len(filter.blockDomains),
		)
	}
}

func TestActionWildcardDNSMasq(t *testing.T) {
	corefile := `filter {
		block wildcard address=/example.com/#
	}`
	filter := NewTestFilter(t, corefile)
	filter.Build()
}

func TestListResolver(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping listresolver test; dns firewalling will cause this to fail")
	}
	http.DefaultTransport = nil
	corefile := `filter {
		listresolver 9.9.9.9
		block list domain https://dbl.oisd.nl/basic/
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
		block list domain https://dbl.oisd.nl/basic/
	}`
	filter := NewTestFilter(t, corefile)
	filter.Build()
	if len(filter.blockDomains) == 0 {
		t.Errorf("expected domains; got none")
	}
}
