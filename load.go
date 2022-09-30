package filter

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"strings"
	"time"
)

// ListLoader contains the means to retrieve a list
type ListLoader interface {
	Load(string) (io.ReadCloser, error)
}

// GetListLoader returns the ListLoader associated with the list's URL scheme
func (a ActionConfig) GetListLoader(uri string) (ListLoader, error) {
	listUrl, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid list URL %q; %w", uri, err)
	}
	if len(listUrl.Scheme) == 0 || len(listUrl.Host) == 0 {
		// URL parsing ambiguity means Parse does not always return an error
		// with an invalid URL. Explicitly check that the Scheme and Host are
		// non-zero values so that they will be usable later.
		return nil, fmt.Errorf("invalid list URL %q; scheme or host empty", uri)
	}
	switch listUrl.Scheme {
	case "file":
		return a.FileLoader, nil
	case "http", "https":
		return a.HTTPLoader, nil
	default:
		return nil, fmt.Errorf(
			"unsupported list URL scheme %q; "+
				"expected 'file', 'http', or 'https'",
			listUrl.Scheme,
		)
	}
}

// FileListLoader retrieves lists from the local filesystem
type FileListLoader struct{}

// Load implements ListLoader
func (FileListLoader) Load(path string) (io.ReadCloser, error) {
	trimmedPath := strings.TrimPrefix(path, "file://")
	file, err := os.Open(trimmedPath)
	if err != nil {
		return nil, fmt.Errorf("error opening list %q; %w", path, err)
	}
	return file, nil
}

// HTTPListLoader retrieves lists from remote sources using HTTP/HTTPS
type HTTPListLoader struct {
	Network    string
	ResolverIP netip.AddrPort
}

// Load implements ListLoader
func (h HTTPListLoader) Load(path string) (io.ReadCloser, error) {
	client := http.Client{}
	if h.ResolverIP.IsValid() {
		dialFunc := func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Second * 5,
			}
			return d.DialContext(ctx, h.Network, h.ResolverIP.String())
		}
		dialerResolver := &net.Resolver{
			PreferGo: true,
			Dial:     dialFunc,
		}
		dialer := &net.Dialer{
			Resolver: dialerResolver,
		}
		dialCtx := func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, addr)
		}
		client.Transport = &http.Transport{
			DialContext: dialCtx,
		}
	}
	resp, err := client.Get(path)
	if err != nil {
		return nil, fmt.Errorf("error fetching list %q; %w", path, err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(
			"an error occurred fetching list %q; %s",
			path,
			resp.Status,
		)
	}
	return resp.Body, nil
}
