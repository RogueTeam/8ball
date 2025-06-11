package old_rpc

import "context"

type SendRawTransactionRequest struct {
	TxAsHex    string `json:"tx_as_hex"`
	DoNotRelay bool   `json:"do_not_relay"` // default false
}

type SendRawTransactionResponse struct {
	Status        string `json:"status"`
	DoubleSpend   bool   `json:"double_spend"`
	FeeTooLow     bool   `json:"fee_too_low"`
	InvalidInput  bool   `json:"invalid_input"`
	InvalidOutput bool   `json:"invalid_output"`
	LowMixin      bool   `json:"low_mixin"`
	NotRct        bool   `json:"not_rct"`
	NotRelayed    bool   `json:"not_relayed"`
	Overspend     bool   `json:"overspend"`
	Reason        string `json:"reason"`
	TooBig        bool   `json:"too_big"`
	Untrusted     bool   `json:"untrusted"`
}

func (c *Client) SendRawTransaction(ctx context.Context, req *SendRawTransactionRequest) (*SendRawTransactionResponse, error) {
	resp := &SendRawTransactionResponse{}
	err := c.Do(ctx, "send_raw_transaction", req, resp)
	return resp, err
}
