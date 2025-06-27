package rpc

import (
	"net/http"
)

// Config holds the configuration of a monero rpc client.
type Config struct {
	// URL including the /json_rpc endpoint
	// Example: http://127.0.0.1:18081/json_rpc
	Url string
	// Custom headers to send
	CustomHeaders map[string]string
	// HTTP Client to use
	Client *http.Client
}
