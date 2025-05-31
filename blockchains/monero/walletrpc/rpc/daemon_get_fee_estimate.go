package rpc

import "context"

type DaemonGetFeeEstimateResponse struct {
	Credits          uint64   `json:"credits"`
	Fee              uint64   `json:"fee"`               // atomic units / byte
	Fees             []uint64 `json:"fees"`              // base fees at different priotities: slow, normal, fast, fastest
	QuantizationMask uint64   `json:"quantization_mask"` // Final fee should be rounded up to an even multiple of this value
	Status           string   `json:"status"`
	TopHash          string   `json:"top_hash"`
	Untrusted        bool     `json:"untrusted"`
}

// Return a fee estimate from the daemon
func (c *Client) DaemonGetFeeEstimate(ctx context.Context) (*DaemonGetBlockCountResponse, error) {
	resp := &DaemonGetBlockCountResponse{}
	err := c.Do(ctx, "get_fee_estimate", nil, resp)
	return resp, err
}
