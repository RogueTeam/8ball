package old_rpc

import (
	"net/http"
)

// Config holds the configuration of a monero rpc client.
type Config struct {
	Address       string
	CustomHeaders map[string]string
	Client        *http.Client
}
