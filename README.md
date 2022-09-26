# filter

[![Go Reference](https://pkg.go.dev/badge/github.com/wranders/coredns-filter.svg)](https://pkg.go.dev/github.com/wranders/coredns-filter)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=wranders_coredns-filter&metric=coverage)](https://sonarcloud.io/summary/overall?id=wranders_coredns-filter)
[![Go Report Card](https://goreportcard.com/badge/github.com/wranders/coredns-filter)](https://goreportcard.com/report/github.com/wranders/coredns-filter)

*filter* - provides domain blocking functionality

## Description

The *filter* plugin is used to block domain name resolution, simliar to
[Pi-hole®](https://github.com/pi-hole/pi-hole).

## Syntax

```nginx
filter {
    ACTION TYPE DATA
}
```

* **ACTION**: `[ allow | block ]` What action to take
* **TYPE**: `[ domain | regex | wildcard ]` What type of **DATA**
* **DATA**:
  * `domain`: A raw domain to match. Subdomains are not matched
  * `regex`: A Go-formatted Regular Expression
  * `wildcard`: Common wildcard formats
    * Generic: `*.example.com`
    * Adblock Plus: `||example.com^`
    * DNSMasq Address: `address=/example.com/#`

```nginx
filter {
    ACTION list TYPE DATA
}
```

* **ACTION**: `[ allow | block ]` What action to take
* **DATA**: Lists of the following data types
  * `domain`: A raw domain to match. Subdomains are not matched
  * `hosts`: A hostsfile formatted list
  * `regex`: A Go-formatted Regular Expression
  * `wildcard`: Common wildcard formats
    * Generic: `*.example.com`
    * Adblock Plus: `||example.com^`
    * DNSMasq Address: `address=/example.com/#`
* **DATA**: A `[ file | http | https ]` URL. Must contain only the **TYPE**
specified.

```nginx
filter {
    response TYPE [ DATA ]
}
```

* **TYPE** `[ address | nodata | null | nxdomain ]` Record type that should be
the response to blocked domains
  * `address`: Only `A` and `AAAA` records are accepted, and only one of each.
  Lowercase record types are accepted.
  * `nodata`: Returns success code but no records
  * `null` (DEFAULT): Returns unspecified address records `A 0.0.0.0` and
  `AAAA ::`
  * `nxdomain`: Returns an `SOA` record for the requested domain.
* **DATA** Only used for `address` responses

```nginx
filter {
    update DURATION
}
```

* **DURATION** (DEFAULT=`24h`): any value accepted by
[`time.ParseDuration`](https://pkg.go.dev/time#ParseDuration)

## Domain Matching

| Directive                         | Description
| :-                                | :-
| `block domain example.com`        | block requests to `example.com` but allow `sub.example.com`
| `block regex .*.example.com`      | block all subdomains of `example.com` but allow requests to `example.com`
| `block wildcard *.example.com`    | block requests to `example.com` and all subdomains

Values for `regex` directives are parsed directly by
[`regexp.Compile`](https://pkg.go.dev/regexp#Compile), so if you're unfamiliar
with Go regular expressions, verify them using
[https://regex101.com/?flavor=golang](https://regex101.com/?flavor=golang).
Complex regular expressions should be loaded from a list instead of inline to
avoid confusing the CoreDNS Corefile parser with symbols.

With how wildcard strings are cleaned and compiled, the following
`block wildcard` directives are identical.

```nginx
filter {
    block wildcard example.com
    block wildcard *.example.com
    block wildcard .*.example.com
    block wildcard ||example.com^
    block wildcard address=/example.com/#
    block wildcard address=/example.com/0.0.0.0
}
```

This flexibility is extended to wildcard lists as well. AdblockPlus and DNSMasq
formats are supported for flexibility and ease of migration from other
solutions. Zone and Unbound configuration files are not supported.

## Examples

```nginx
# Use an aggregated block list, but allow `vortex.data.microsoft.com` for XBox
# Live achievements. Respond to blocked domains with `A 0.0.0.0` and `AAAA ::`
# records. Update list every 24 hours.

filter {
    block list domain https://dbl.oisd.nl/basic/
    allow domain vortex.data.microsoft.com
}
```

```nginx
# Block all request except domains explicitly allowed in the list on the
# filesystem. Forward blocked requests to an internal web server to display a
# block page and log unapproved sites. Do not automatically update lists. Only a
# restart of CoreDNS will refresh the list's contents.

filter {
    block regex .*
    allow list domains file:///etc/coredns/whitelist
    response address A 10.0.1.50 AAAA 2001:db8:abcd:0012::ffff:0a00:0132
    update 0
}
```

## Buliding

Clone the [coredns](https://github.com/coredns/coredns) repository and change
into it's directory.

```sh
git clone https://github.com/coredns/coredns.git
```

```sh
cd coredns
```

Fetch the plugin and add it to `coredns`'s `go.mod` file:

```sh
go get -u github.com/wranders/coredns-filter
```

Update `plugin.cfg` in the root of the directory. The `filter` declaration
should be inserted before `cache` so that updates to the `filter` are applied
immediately.

```sh
# Using sed
sed -i '/^cache:cache/i filter:github.com/wranders/coredns-filter' plugin.cfg
```

```powershell
# Using Powershell
(Get-Content plugin.cfg).`
Replace("cache:cache", "filter:github.com/wranders/coredns-filter`ncache:cache") | `
Set-Content -Path plugin.cfg
```

Generate the plugin files:

```sh
go generate
```

And build:

```sh
go build
```

The `coredns` binary will be in the root of the project directory, unless
otherwise specified by the `-o` flag.

## License

This plugin is licensed under the MIT license. See [LICENSE](./LICENSE) for more
information.
