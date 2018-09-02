package gahttp

import (
	"crypto/tls"
	"net/http"
)

// NewDefaultClient returns the default HTTP client
func NewDefaultClient() *http.Client {
	return &http.Client{}
}

// ClientOptions are a bitmask of options for HTTP clients
type ClientOptions int

const (
	// Don't follow redirects
	NoRedirects ClientOptions = 1 << iota

	// Skip verification of TLS certificates
	SkipVerify
)

// NewClient returns a new client with the specified options
func NewClient(opts ClientOptions) *http.Client {

	transport := &http.Transport{}

	if opts&SkipVerify > 0 {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client := &http.Client{
		Transport: transport,
	}

	if opts&NoRedirects > 0 {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return client
}
