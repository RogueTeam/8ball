package rpc

import "context"

type DaemonGetBlockCountResponse struct {
	Count     uint64 `json:"count"` // the number of blocks not the zero based height
	Status    string `json:"status"`
	Untrusted bool   `json:"untrusted"`
}

// Return the current block count known to the daemon
func (c *Client) DaemonGetBlockCount(ctx context.Context) (*DaemonGetBlockCountResponse, error) {
	resp := &DaemonGetBlockCountResponse{}
	err := c.Do(ctx, "get_block_count", nil, resp)
	return resp, err
}
