// Command healthprobe exits 0 when a target is healthy and 1 otherwise.
//
//	healthprobe http://127.0.0.1:8080/healthz
//	healthprobe -contains '"status":"ok"' http://127.0.0.1:8080/healthz
//	healthprobe tcp://127.0.0.1:5432
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PrinceAdvant/healthprobe"
)

type headers []string

func (h *headers) String() string     { return strings.Join(*h, ", ") }
func (h *headers) Set(v string) error { *h = append(*h, v); return nil }

func main() {
	timeout := flag.Duration("timeout", 3*time.Second, "timeout for the check")
	contains := flag.String("contains", "", "text the HTTP response body must contain")
	minS := flag.Int("min-status", 200, "lowest accepted HTTP status")
	maxS := flag.Int("max-status", 399, "highest accepted HTTP status")
	quiet := flag.Bool("q", false, "print nothing on success")
	var hs headers
	flag.Var(&hs, "H", "extra request header, for example -H 'Host: app.local' (repeatable)")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: healthprobe [flags] <http(s)://... | tcp://host:port>")
		os.Exit(2)
	}
	h := http.Header{}
	for _, v := range hs {
		k, val, ok := strings.Cut(v, ":")
		if ok {
			h.Add(strings.TrimSpace(k), strings.TrimSpace(val))
		}
	}
	err := healthprobe.Check(context.Background(), flag.Arg(0), healthprobe.Options{
		Timeout: *timeout, Contains: *contains, MinStatus: *minS, MaxStatus: *maxS, Header: h,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "unhealthy:", err)
		os.Exit(1)
	}
	if !*quiet {
		fmt.Println("healthy")
	}
}
