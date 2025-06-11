package old_rpc

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"anarchy.ttfm.onion/gateway/internal/walletrpc/old_rpc/json"
)

type Client struct {
	httpcl  *http.Client
	addr    string
	headers map[string]string
}

// New returns a new monerod-rpc client.
func New(cfg Config) *Client {
	cl := &Client{
		addr:    cfg.Address,
		headers: cfg.CustomHeaders,
	}
	if cfg.Client == nil {
		cl.httpcl = http.DefaultClient
	} else {
		cl.httpcl = cfg.Client
	}
	return cl
}

func (c *Client) Do(ctx context.Context, method string, in, out interface{}) error {
	payload, err := json.EncodeClientRequest(in)
	if err != nil {
		return err
	}
	// fmt.Printf("payload: %s\n", string(payload))
	addr := c.addr + "/" + method
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, addr, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.headers != nil {
		for k, v := range c.headers {
			req.Header.Set(k, v)
		}
	}
	resp, err := c.httpcl.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status %v", resp.StatusCode)
	}
	defer resp.Body.Close()

	// this is done to catch any monero related errors if we are not expecting any
	// data back
	if out == nil {
		v := &json.EmptyResponse{}
		return json.DecodeClientResponse(resp.Body, v)
	}
	return json.DecodeClientResponse(resp.Body, out)
}
