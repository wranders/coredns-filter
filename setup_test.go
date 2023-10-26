package filter

import (
	"errors"
	"testing"

	"github.com/coredns/caddy"
)

type TestSetup struct {
	Name     string
	Corefile string
	WantErr  bool
}

func RunSetupTest(t *testing.T, test TestSetup) {
	t.Run(test.Name, func(t *testing.T) {
		controller := caddy.NewTestController("dns", test.Corefile)
		if err := setup(controller); (err != nil) != test.WantErr {
			t.Errorf(
				"error: setup: %v, wanterr: %t",
				err,
				test.WantErr,
			)
		}
	})
}

func TestSetupMain(t *testing.T) {
	tests := []TestSetup{
		{
			"default",
			`filter`,
			false,
		},
		{
			"unknown directive",
			`filter {
				noop
			}`,
			true,
		},
	}
	for _, test := range tests {
		RunSetupTest(t, test)
	}
}

func TestSetupActionParse(t *testing.T) {
	tests := []TestSetup{
		{
			"action no type",
			`filter {
				allow
			}`,
			true,
		},
		{
			"action unknown type",
			`filter {
				allow noop
			}`,
			true,
		},
		{
			"action list no type",
			`filter {
				allow list
			}`,
			true,
		},
		{
			"action list unknown type",
			`filter {
				allow list noop
			}`,
			true,
		},
	}
	for _, test := range tests {
		RunSetupTest(t, test)
	}
}

func TestSetupUpdate(t *testing.T) {
	tests := []TestSetup{
		{
			"update no interval",
			`filter {
				update
			}`,
			true,
		},
		{
			"update 6 hour",
			`filter {
				update 6h
			}`,
			false,
		},
		{
			"update invalid interval",
			`filter {
				update -999
			}`,
			true,
		},
	}
	for _, test := range tests {
		RunSetupTest(t, test)
	}
}

func TestSetupResponseAddress(t *testing.T) {
	tests := []TestSetup{
		{
			"response address a record no address",
			`filter {
				response address a
			}`,
			true,
		},
		{
			"response address a record invalid address",
			`filter {
				response address a 192.168.1
			}`,
			true,
		},
		{
			"response address a record",
			`filter {
				response address a 192.168.1.1
			}`,
			false,
		},
		{
			"response address a record ipv6 mismatch",
			`filter {
				response address a 2001:db8::0:1
			}`,
			true,
		},
		{
			"response address a record uppercase",
			`filter {
				response address A 192.168.1.1
			}`,
			false,
		},
		{
			"response address aaaa record no address",
			`filter {
				response address aaaa
			}`,
			true,
		},
		{
			"response address a record invalid address",
			`filter {
				response address aaaa 2001:db8:0:1
			}`,
			true,
		},
		{
			"response address aaaa record",
			`filter {
				response address aaaa 2001:db8::0:1
			}`,
			false,
		},
		{
			"response address aaaa record ipv4 mismatch",
			`filter {
				response address aaaa 192.168.1.1
			}`,
			true,
		},
		{
			"response address aaaa record uppercase",
			`filter {
				response address AAAA 2001:db8::0:1
			}`,
			false,
		},
		{
			"response address a+aaaa records",
			`filter {
				response address a 192.168.1.1 aaaa 2001:db8::0:1
			}`,
			false,
		},
		{
			"response address a+aaaa records uppercase",
			`filter {
				response address A 192.168.1.1 AAAA 2001:db8::0:1
			}`,
			false,
		},
		{
			"response address aaaa+a records",
			`filter {
				response address aaaa 2001:db8::0:1 a 192.168.1.1
			}`,
			false,
		},
		{
			"response address aaaa+a records uppercase",
			`filter {
				response address AAAA 2001:db8:0::1 A 192.168.1.1
			}`,
			false,
		},
		{
			"response address invalid second record type",
			`filter {
				response address a 192.168.1.1 cname example.com
			}`,
			true,
		},
		{
			"response address missing second record address",
			`filter {
				response address a 192.168.1.1 aaaa
			}`,
			true,
		},
		{
			"response address invalid second record address",
			`filter {
				response address a 192.168.1.1 aaaa 2001:db8:0:1
			}`,
			true,
		},
		{
			"response address duplicate records",
			`filter {
				response address a 192.168.1.1 a 192.168.1.2
			}`,
			true,
		},
		{
			"response address second record type mismatch aaaa+ipv4",
			`filter {
				response address a 192.168.1.1 aaaa 192.168.1.2
			}`,
			true,
		},
		{
			"response address second record type mismatch a+ipv4",
			`filter {
				response address aaaa 2001:db8::0:1 a 2001:db8::0:2
			}`,
			true,
		},
		{
			"response address more than two records",
			`filter {
				response address aaaa 2001:db8::0:1 a 192.168.1.1 a 192.168.1.2
			}`,
			true,
		},
	}
	for _, test := range tests {
		RunSetupTest(t, test)
	}
}

func TestSetupListResolver(t *testing.T) {
	tests := []TestSetup{
		{
			"listresolver none specified",
			`filter {
				listresolver
			}`,
			true,
		},
		{
			"listresolver nonsensical",
			`filter {
				listresolver noop
			}`,
			true,
		},
		{
			"listresolver ip colon no port",
			`filter {
				listresolver 9.9.9.9:
			}`,
			true,
		},
		{
			"listresolver quad9 default",
			`filter {
				listresolver 9.9.9.9
			}`,
			false,
		},
		{
			"listresolver quad9 dns",
			`filter {
				listresolver dns://9.9.9.9
			}`,
			false,
		},
		{
			"listresolver quad9 tls without server name",
			`filter {
				listresolver tls://9.9.9.9
			}`,
			true,
		},
		{
			"listresolver quad9 tls without server name",
			`filter {
				listresolver tls://9.9.9.9 dns.quad9.net
			}`,
			false,
		},
		{
			"listresolver quad9 unsupported transport",
			`filter {
				listresolver https://9.9.9.9
			}`,
			true,
		},
		{
			"listresolver quad9 domain name",
			`filter {
				listresolver tls://dns.quad9.net
			}`,
			true,
		},
	}
	for _, test := range tests {
		RunSetupTest(t, test)
	}
}

func TestSetupExpectedEOL(t *testing.T) {
	corefile := `filter {
		allow domain example.com noop
	}`
	t.Run("action expected eol", func(t *testing.T) {
		controller := caddy.NewTestController("dns", corefile)
		err := setup(controller)
		t.Logf("%s", err)
		var acceptErr errorExpectedEOL
		if !errors.As(err, &acceptErr) {
			t.Errorf(
				"expected eol error type; got %T",
				err,
			)
		}
	})
}
