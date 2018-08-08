package main

import (
	"net"
	"net/url"
	"time"
)

func runHealthCheck() {
	url, _ := url.Parse(BootstrapHost)
	address := url.Host
	if url.Port() == "" {
		if url.Scheme == "https" {
			address = address + ":443"
		} else if url.Scheme == "http" {
			address = address + ":80"
		}
	}
	failures := 0
	maxFailures := 10
	for range time.Tick(1 * time.Minute) {
		conn, err := net.DialTimeout("tcp", url.Host, 5*time.Second)
		if err != nil {
			failures++
			if failures > maxFailures {
				handleHealthFailure()
			}
			continue
		}
		conn.Close()
		failures = 0
	}
}
