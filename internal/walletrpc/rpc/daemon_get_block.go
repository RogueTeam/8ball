package rpc

import "context"

type DaemonGetBlockResponse struct {
	Blob        string      `json:"blob"`
	BlockHeader BlockHeader `json:"block_header"`
	Credits     uint64      `json:"credits"`
	Json        string      `json:"json"`          // json formatted details
	MinerTxHash string      `json:"miner_tx_hash"` // coinbase tx
	Status      string      `json:"status"`
	TopHash     string      `json:"top_hash"`
	TxHashes    []string    `json:"tx_hashes"` // non-coinbase txs
	Untrusted   bool        `json:"untrusted"`
}

// Inputs (pick height or hash):
type DaemonGetBlockRequest struct {
	// The block's height.
	Height uint64 `json:"height"`
	// The block's sha256 hash.
	Hash string `json:"hash"`
	// Optional; defaults to false. Add PoW hash to block_header response.
	FillPowHash bool `json:"fill_pow_hash"`
}

// Return block info from the daemon given the hash
func (c *Client) DaemonGetBlock(ctx context.Context, req *DaemonGetBlockRequest) (*DaemonGetBlockResponse, error) {
	resp := &DaemonGetBlockResponse{}
	err := c.Do(ctx, "get_block", req, resp)
	return resp, err
}
