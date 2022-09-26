package filter

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// ListLoadFunc represents a function used to fetch an action list
type ListLoadFunc func(string) (io.ReadCloser, error)

// GetListLoadFunc returns the function used to load an action list based on the
// scheme of the URL
func GetListLoadFunc(uri string) (ListLoadFunc, error) {
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
		return LoadFile, nil
	case "http", "https":
		return LoadHttp, nil
	default:
		return nil, fmt.Errorf(
			"unsupported list URL scheme %q; "+
				"expected 'file', 'http', or 'https'",
			listUrl.Scheme,
		)
	}
}

// LoadFile implements ListLoadFunc. Fetches the contents of a local file
func LoadFile(path string) (io.ReadCloser, error) {
	trimmedPath := strings.TrimPrefix(path, "file://")
	file, err := os.Open(trimmedPath)
	if err != nil {
		return nil, fmt.Errorf("error opening list %q; %w", path, err)
	}
	return file, nil
}

// LoadHttp implements ListLoadFunc. Fetches the contents of a remote file over
// HTTP/S
func LoadHttp(path string) (io.ReadCloser, error) {
	resp, err := http.Get(path)
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
