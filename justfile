
[doc("""
Run Tests
    Pass '-short' to skip tests that require external DNS services.
    Useful if outbound DNS requests are blocked by a firewall.
""")]
[arg('short', pattern='|-short')]
test *short:
    go test {{short}} -coverprofile='coverage.out'

[doc("""
Run Tests and View Coverage
    Pass '-short' to skip tests that require external DNS services.
    Useful if outbound DNS requests are blocked by a firewall.
""")]
[arg('short', pattern='-short')]
coverage *short: (test short)
    go tool cover -html='coverage.out'

# Launch pkgsite to view documentation (requires pkgsite)
pkgsite:
    pkgsite -list=false
