package rpc

import "context"

//////////////////////////////////////////
// THIS API ONLY WORKS ON REGTEST (SIMNET)
//////////////////////////////////////////

// Start monerod with these flags included: --regtest --offline --fixed-difficulty 1.
// --offline ensures that the node does not connect to the main network and learn
//   of its latest chaintip.
// --fixed-difficulty keeps the difficulty constant, allowing a large number o
//   blocks to be generated quickly.

type DemonRegtestGenerateResponse struct {
	Status    string   `json:"status"`
	Blocks    []string `json:"blocks"`
	Height    uint64   `json:"height"`
	Untrusted bool     `json:"untrusted"`
}

type DemonRegtestGenerateRequest struct {
	AmountOfBlocks uint64 `json:"amount_of_blocks"`     // We will default to 1 here
	WalletAddress  string `json:"wallet_address"`       // Address to mine to - a wallet Primary address Account 0 Sub [0]
	PrevBlock      string `json:"prev_block,omitempty"` // I think is an old param - not sure what is for
	StartingNonce  uint64 `json:"starting_nonce"`       // Defaults to 0 here if not specified
}

// Mines 1 or more blocks on nettype == regtest
func (c *Client) DaemonRegtestGenerate(ctx context.Context, req *DemonRegtestGenerateRequest) (*DemonRegtestGenerateResponse, error) {
	resp := &DemonRegtestGenerateResponse{}
	if req.AmountOfBlocks == 0 {
		req.AmountOfBlocks = 1
	}
	err := c.Do(ctx, "generateblocks", req, resp)
	return resp, err
}
