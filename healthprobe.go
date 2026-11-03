// Package healthprobe performs a single HTTP or TCP health check and reports
// success or failure. It is designed to be compiled into a tiny static binary
// that can be used in Docker HEALTHCHECK instructions for images that have no
// shell, curl or wget, such as distroless and scratch images.
package healthprobe

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Options controls a probe.
type Options struct {
	// Timeout for the whole check. Defaults to 3 seconds.
	Timeout time.Duration
	// MinStatus and MaxStatus define the accepted HTTP status range.
	// Defaults to 200..399.
	MinStatus, MaxStatus int
	// Contains, if set, must appear in the first 64 KiB of the response body.
	Contains string
	// Header values added to HTTP requests, for example a Host header.
	Header http.Header
}

func (o *Options) defaults() {
	if o.Timeout <= 0 {
		o.Timeout = 3 * time.Second
	}
	if o.MinStatus == 0 {
		o.MinStatus = 200
	}
	if o.MaxStatus == 0 {
		o.MaxStatus = 399
	}
}

// Check runs one probe against target, which is either an http(s) URL or
// tcp://host:port. It returns nil when the target is healthy.
func Check(ctx context.Context, target string, o Options) error {
	o.defaults()
	ctx, cancel := context.WithTimeout(ctx, o.Timeout)
	defer cancel()
	u, err := url.Parse(target)
	if err != nil {
		return fmt.Errorf("parse target: %w", err)
	}
	switch u.Scheme {
	case "tcp":
		var d net.Dialer
		c, err := d.DialContext(ctx, "tcp", u.Host)
		if err != nil {
			return err
		}
		return c.Close()
	case "http", "https":
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
		if err != nil {
			return err
		}
		for k, vs := range o.Header {
			for _, v := range vs {
				if strings.EqualFold(k, "Host") {
					req.Host = v
					continue
				}
				req.Header.Add(k, v)
			}
		}
		client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode < o.MinStatus || resp.StatusCode > o.MaxStatus {
			return fmt.Errorf("status %d outside %d..%d", resp.StatusCode, o.MinStatus, o.MaxStatus)
		}
		if o.Contains != "" {
			b, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
			if !strings.Contains(string(b), o.Contains) {
				return errors.New("response body does not contain expected text")
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported scheme %q, use http, https or tcp", u.Scheme)
	}
}
