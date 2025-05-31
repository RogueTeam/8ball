package rpc

import "context"

type DaemonGetBlockHeaderResponse struct {
	BlockHeader BlockHeader `json:"block_header"`
	Credits     uint64      `json:"credits"`
	Status      string      `json:"status"`
	TopHash     string      `json:"top_hash"`
	Untrusted   bool        `json:"untrusted"`
}

type DaemonGetBlockHeaderByHashRequest struct {
	// The block's sha256 hash.
	Hash string `json:"hash"`
	// Optional; defaults to false. Add PoW hash to block_header response.
	FillPowHash bool `json:"fill_pow_hash"`
}

// Return block header info from the daemon given the hash
func (c *Client) DaemonGetBlockHeaderByHash(ctx context.Context, req *DaemonGetBlockHeaderByHashRequest) (*DaemonGetBlockHeaderResponse, error) {
	resp := &DaemonGetBlockHeaderResponse{}
	err := c.Do(ctx, "get_block_header_by_hash", req, resp)
	return resp, err
}

type DaemonGetBlockHeaderByHeightRequest struct {
	// The block's height.
	Height uint64 `json:"height"`
	// Optional; defaults to false. Add PoW hash to block_header response.
	FillPowHash bool `json:"fill_pow_hash"`
}

// Return block header info from the daemon given the height
func (c *Client) DaemonGetBlockHeaderByHeight(ctx context.Context, req *DaemonGetBlockHeaderByHeightRequest) (*DaemonGetBlockHeaderResponse, error) {
	resp := &DaemonGetBlockHeaderResponse{}
	err := c.Do(ctx, "get_block_header_by_height", req, resp)
	return resp, err
}
