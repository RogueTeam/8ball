package rpc

import "context"

type DaemonGetLastBlockHeaderRequest struct {
	// Optional; defaults to false. Add PoW hash to block_header response.
	FillPowHash bool `json:"fill_pow_hash"`
}

// Return latest block header info from the daemon
func (c *Client) DaemonGetLastBlockHeader(ctx context.Context, req *DaemonGetLastBlockHeaderRequest) (*DaemonGetBlockHeaderResponse, error) {
	resp := &DaemonGetBlockHeaderResponse{}
	err := c.Do(ctx, "get_last_block_header", req, resp)
	return resp, err
}
