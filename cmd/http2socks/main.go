// Translate HTTP proxy requests to SOCKS5 oness
package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"net"
	"net/http"
	"os"

	goproxy "github.com/elazarl/goproxy"
	"golang.org/x/net/proxy"
)

var config struct {
	// Address to listen to
	address string
	// Socks
	socks string
	// Insecure TLS
	insecureTls bool
}

func init() {
	set := flag.NewFlagSet("http2socks", flag.ExitOnError)
	set.StringVar(&config.address, "listen", "127.0.0.1:8080", "Listen address")
	set.StringVar(&config.socks, "socks5", "127.0.0.1:9050", "SOCKS5 Address")
	set.BoolVar(&config.insecureTls, "insecure-tls", false, "Use insecure TLS")

	err := set.Parse(os.Args[1:])
	if err != nil {
		log.Fatal("failed to parse flags: %w", err)
	}
}

func main() {
	dialer, err := proxy.SOCKS5("tcp", config.socks, nil, nil)
	if err != nil {
		log.Fatal("Failed to prepare SOCKS5 Dialer")
	}

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true
	proxy.ConnectDial = dialer.Dial
	proxy.ConnectDialWithReq = func(req *http.Request, network, addr string) (net.Conn, error) {
		return dialer.Dial(network, addr)
	}
	proxy.Tr.Dial = dialer.Dial
	proxy.Tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.Dial(network, addr)
	}
	proxy.Tr.DialTLSContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := dialer.Dial(network, addr)
		if err == nil {
			return tls.Client(conn, &tls.Config{InsecureSkipVerify: config.insecureTls}), nil
		}
		return nil, err
	}

	err = http.ListenAndServe(config.address, proxy)
	if err != nil {
		log.Fatal(err)
	}
}
