package filter

import (
	"errors"
	"strings"
	"testing"

	"github.com/coredns/caddy"
)

type testSetup struct {
	Name     string
	Corefile string
	WantErr  bool
}

func RunSetupTests(t *testing.T, tests []testSetup) {
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			controller := caddy.NewTestController("dns", test.Corefile)
			if err := setup(controller); (err != nil) != test.WantErr {
				t.Errorf(
					"Error: setup() error = %v, WantErr %v",
					err,
					test.WantErr,
				)
			}
		})
	}
}

func TestSetup(t *testing.T) {
	tests := []testSetup{
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

	RunSetupTests(t, tests)
}

func TestAllow(t *testing.T) {
	tests := []testSetup{
		{
			"allow no type",
			`filter {
				allow
			}`,
			true,
		},
		{
			"allow domain none",
			`filter {
				allow domain
			}`,
			true,
		},
		{
			"allow domain",
			`filter {
				allow domain example.com
			}`,
			false,
		},
		{
			"allow domain duplicate",
			`filter {
				allow domain example.com
				allow domain example.com
			}`,
			false,
		},
		{
			"allow regex no expression",
			`filter {
				allow regex
			}`,
			true,
		},
		{
			"allow regex",
			`filter {
				allow regex .*.example.com
			}`,
			false,
		},
		{
			"allow regex invalid expression",
			`filter {
				allow regex *.example.com
			}`,
			true,
		},
		{
			"allow regex duplicate",
			`filter {
				allow regex .*.example.com
				allow regex .*.example.com
			}`,
			false,
		},
		{
			"allow wildcard no url",
			`filter {
				allow wildcard
			}`,
			true,
		},
		{
			"allow wildcard",
			`filter {
				allow wildcard example.com
			}`,
			false,
		},
		{
			"allow wildcard as common regexp",
			`filter {
				allow wildcard *.example.com
			}`,
			false,
		},
		{
			"allow wildcard invalid character",
			`filter {
				allow wildcard ex[ample.com
			}`,
			true,
		},
		{
			"allow no type",
			`filter {
				allow wildcard example.com
				allow wildcard *.example.com
			}`,
			false,
		},
		{
			"allow unknown directive",
			`filter {
				allow noop
			}`,
			true,
		},
	}

	RunSetupTests(t, tests)
}

func TestBlock(t *testing.T) {
	tests := []testSetup{
		{
			"block no type",
			`filter {
				block
			}`,
			true,
		},
		{
			"block domain none",
			`filter {
				block domain
			}`,
			true,
		},
		{
			"block domain",
			`filter {
				block domain example.com
			}`,
			false,
		},
		{
			"block domain duplicate",
			`filter {
				block domain example.com
				block domain example.com
			}`,
			false,
		},
		{
			"block regex no expression",
			`filter {
				block regex
			}`,
			true,
		},
		{
			"block regex",
			`filter {
				block regex .*.example.com
			}`,
			false,
		},
		{
			"block regex invalid expression",
			`filter {
				block regex *.example.com
			}`,
			true,
		},
		{
			"block regex duplicate",
			`filter {
				block regex .*.example.com
				block regex .*.example.com
			}`,
			false,
		},
		{
			"block wildcard no url",
			`filter {
				block wildcard
			}`,
			true,
		},
		{
			"block wildcard",
			`filter {
				block wildcard example.com
			}`,
			false,
		},
		{
			"block wildcard as common regexp",
			`filter {
				block wildcard *.example.com
			}`,
			false,
		},
		{
			"block wildcard invalid character",
			`filter {
				block wildcard ex[ample.com
			}`,
			true,
		},
		{
			"block no type",
			`filter {
				block wildcard example.com
				block wildcard *.example.com
			}`,
			false,
		},
		{
			"block unknown directive",
			`filter {
				block noop
			}`,
			true,
		},
	}

	RunSetupTests(t, tests)
}

func TestAllowList(t *testing.T) {
	tests := []testSetup{
		{
			"allow list no type",
			`filter {
				allow list
			}`,
			true,
		},
		{
			"allow list unknown type",
			`filter {
				allow list noop
			}`,
			true,
		},
		{
			"allow list domain no url",
			`filter {
				allow list domain
			}`,
			true,
		},
		{
			"allow list domain no url",
			`filter {
				allow list domain
			}`,
			true,
		},
		{
			"allow list domain invalid url",
			`filter {
				allow list domain "file://invalid url"
			}`,
			true,
		},
		{
			"allow list domain invalid format",
			`filter {
				allow list domain "file://.testdata/unknown.list"
			}`,
			false,
		},
		{
			"allow list domain no scheme",
			`filter {
				allow list domain noop
			}`,
			true,
		},
		{
			"allow list domain invalid scheme",
			`filter {
				allow list domain scheme://noop
			}`,
			true,
		},
		{
			"allow list domain file",
			`filter {
				allow list domain file://.testdata/domain.list
			}`,
			false,
		},
		{
			"allow list domain duplicate",
			`filter {
				allow list domain file://.testdata/domain.list
				allow list domain file://.testdata/domain.list
			}`,
			false,
		},
		{
			"allow list domain http",
			`filter {
				allow list domain http://dbl.oisd.nl/basic/
			}`,
			false,
		},
		{
			"allow list domain https",
			`filter {
				allow list domain https://dbl.oisd.nl/basic/
			}`,
			false,
		},
		{
			"allow list hosts no url",
			`filter {
				allow list hosts
			}`,
			true,
		},
		{
			"allow list hosts invalid url",
			`filter {
				allow list hosts "file://invalid url"
			}`,
			true,
		},
		{
			"allow list hosts from file",
			`filter {
				allow list hosts file://.testdata/hosts.list
			}`,
			false,
		},
		{
			"allow list hosts duplicate",
			`filter {
				allow list hosts file://.testdata/hosts.list
				allow list hosts file://.testdata/hosts.list
			}`,
			false,
		},
		{
			"allow list regex no url",
			`filter {
				allow list regex
			}`,
			true,
		},
		{
			"allow list regex invalid url",
			`filter {
				allow list regex "file://invalid url"
			}`,
			true,
		},
		{
			"allow list regex from file",
			`filter {
				allow list regex file://.testdata/regex.list
			}`,
			false,
		},
		{
			"allow list regex duplicate",
			`filter {
				allow list regex file://.testdata/regex.list
				allow list regex file://.testdata/regex.list
			}`,
			false,
		}, {
			"allow list wildcard no url",
			`filter {
				allow list wildcard
			}`,
			true,
		},
		{
			"allow list wildcard invalid url",
			`filter {
				allow list wildcard "file://invalid url"
			}`,
			true,
		},
		{
			"allow list wildcard from file",
			`filter {
				allow list wildcard file://.testdata/regex.list
			}`,
			false,
		},
		{
			"allow list wildcard duplicate",
			`filter {
				allow list wildcard file://.testdata/regex.list
				allow list wildcard file://.testdata/regex.list
			}`,
			false,
		},
	}

	RunSetupTests(t, tests)
}

func TestBlockList(t *testing.T) {
	tests := []testSetup{
		{
			"block list no type",
			`filter {
				block list
			}`,
			true,
		},
		{
			"block list unknown type",
			`filter {
				block list noop
			}`,
			true,
		},
		{
			"block list domain no url",
			`filter {
				block list domain
			}`,
			true,
		},
		{
			"block list domain no url",
			`filter {
				block list domain
			}`,
			true,
		},
		{
			"block list domain invalid url",
			`filter {
				block list domain "file://invalid url"
			}`,
			true,
		},
		{
			"block list domain invalid format",
			`filter {
				block list domain "file://.testdata/unknown.list"
			}`,
			false,
		},
		{
			"block list domain no scheme",
			`filter {
				block list domain noop
			}`,
			true,
		},
		{
			"block list domain invalid scheme",
			`filter {
				block list domain scheme://noop
			}`,
			true,
		},
		{
			"block list domain file",
			`filter {
				block list domain file://.testdata/domain.list
			}`,
			false,
		},
		{
			"block list domain duplicate",
			`filter {
				block list domain file://.testdata/domain.list
				block list domain file://.testdata/domain.list
			}`,
			false,
		},
		{
			"block list domain http",
			`filter {
				block list domain http://dbl.oisd.nl/basic/
			}`,
			false,
		},
		{
			"block list domain https",
			`filter {
				block list domain https://dbl.oisd.nl/basic/
			}`,
			false,
		},
		{
			"block list hosts no url",
			`filter {
				block list hosts
			}`,
			true,
		},
		{
			"block list hosts invalid url",
			`filter {
				block list hosts "file://invalid url"
			}`,
			true,
		},
		{
			"block list hosts from file",
			`filter {
				block list hosts file://.testdata/hosts.list
			}`,
			false,
		},
		{
			"block list hosts duplicate",
			`filter {
				block list hosts file://.testdata/hosts.list
				block list hosts file://.testdata/hosts.list
			}`,
			false,
		},
		{
			"block list regex no url",
			`filter {
				block list regex
			}`,
			true,
		},
		{
			"block list regex invalid url",
			`filter {
				block list regex "file://invalid url"
			}`,
			true,
		},
		{
			"block list regex from file",
			`filter {
				block list regex file://.testdata/regex.list
			}`,
			false,
		},
		{
			"block list regex duplicate",
			`filter {
				block list regex file://.testdata/regex.list
				block list regex file://.testdata/regex.list
			}`,
			false,
		}, {
			"block list wildcard no url",
			`filter {
				block list wildcard
			}`,
			true,
		},
		{
			"block list wildcard invalid url",
			`filter {
				block list wildcard "file://invalid url"
			}`,
			true,
		},
		{
			"block list wildcard from file",
			`filter {
				block list wildcard file://.testdata/regex.list
			}`,
			false,
		},
		{
			"block list wildcard duplicate",
			`filter {
				block list wildcard file://.testdata/regex.list
				block list wildcard file://.testdata/regex.list
			}`,
			false,
		},
	}

	RunSetupTests(t, tests)
}

func TestResponse(t *testing.T) {
	tests := []testSetup{
		{
			"response no type",
			`filter {
				response
			}`,
			true,
		},
		{
			"response unknown type",
			`filter {
				response noop
			}`,
			true,
		},
		{
			"response nodata",
			`filter {
				response nodata
			}`,
			false,
		},
		{
			"response null",
			`filter {
				response null
			}`,
			false,
		},
		{
			"response nxdomain",
			`filter {
				response nxdomain
			}`,
			false,
		},
		{
			"response address no records",
			`filter {
				response address
			}`,
			true,
		},
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

	RunSetupTests(t, tests)
}

func TestUpdate(t *testing.T) {
	tests := []testSetup{
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

	RunSetupTests(t, tests)
}

func TestParserEOLError(t *testing.T) {
	tests := []struct {
		Name     string
		Corefile string
	}{
		{
			"allow domain extra token",
			`filter {
				allow domain google.com noop
			}`,
		},
		{
			"allow regex extra token",
			`filter {
				allow regex .* noop
			}`,
		},
		{
			"allow wildcard extra token",
			`filter {
				allow wildcard google.com noop
			}`,
		},
		{
			"allow list domain extra token",
			`filter {
				allow list domain file://domain.list noop
			}`,
		},
		{
			"allow list regex extra token",
			`filter {
				allow list regex file://domain.list noop
			}`,
		},
		{
			"allow list wildcard extra token",
			`filter {
				allow list regex file://domain.list noop
			}`,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			controller := caddy.NewTestController("dns", test.Corefile)
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
}

func TestParserEOLErrorText(t *testing.T) {
	corefile := `filter {
		allow domain google.com noop
	}`
	expected := `unexpected token(s): ["noop"]; expected end of line`
	controller := caddy.NewTestController("dns", corefile)
	err := setup(controller)
	if err == nil {
		t.Error("an error was expected and not emitted")
	}
	if strings.Compare(err.Error(), expected) != 0 {
		t.Errorf("unexpected error text `%s`; expected `%s`", err, expected)
	}
}
