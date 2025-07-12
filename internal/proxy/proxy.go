package proxy

import (
	"net/http/httputil"
	"net/url"
)

// New creates a new reverse proxy to the given target base URL.
func New(targetBase string) (*httputil.ReverseProxy, error) {
	target, err := url.Parse(targetBase)
	if err != nil {
		return nil, err
	}
	return httputil.NewSingleHostReverseProxy(target), nil
}
