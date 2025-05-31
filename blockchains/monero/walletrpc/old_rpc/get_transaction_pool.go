package old_rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

type SpentKeyImage struct {
	IdHash   string   `json:"id_hash"`
	TxHashes []string `json:"tx_hashes"`
}

type Transaction struct {
	BlobSize           uint64 `json:"blob_size"`
	DoNotRelay         bool   `json:"do_not_relay"`
	DoubleSpendSeen    bool   `json:"double_spend_seen"`
	Fee                uint64 `json:"fee"`
	IdHash             string `json:"id_hash"`
	KeptByBlock        bool   `json:"kept_by_block"`
	LastFailedHeight   uint64 `json:"last_failed_height"`
	LastFailedIdHash   string `json:"last_failed_id_hash"`
	LastRelayedTime    uint64 `json:"last_relayed_time"`
	MaxUsedBlockHeight uint64 `json:"max_used_block_height"`
	MaxUsedBlockIdHash string `json:"max_used_block_id_hash"`
	ReceiveTime        uint64 `json:"receive_time"`
	Relayed            bool   `json:"relayed"`
	TxBlob             string `json:"tx_blob"`
	TxJson             string `json:"tx_json"`
	// Additional tx details:
	// - sent as a raw json string in the 'tx_json' field which we need to further
	//   decode into JsonDetails
	// - Discards all the 'rctsig_prunable' info as probably not useful for us.
	JsonDetails JsonDetails
}

type GetTransactionPoolResponse struct {
	Status         string          `json:"status"`
	Credits        uint64          `json:"credits"`
	SpentKeyImages []SpentKeyImage `json:"spent_key_images"`
	Transactions   []Transaction   `json:"transactions"`
	Untrusted      bool            `json:"untrusted"`
}

// Get transaction pool (mempool) info.
func (c *Client) GetTransactionPool(ctx context.Context) (*GetTransactionPoolResponse, error) {
	resp := &GetTransactionPoolResponse{}
	err := c.Do(ctx, "get_transaction_pool", nil, resp)
	if err != nil {
		return nil, err
	}
	if resp.Transactions != nil {
		// transaction pool not empty
		err = decodeTxJson(resp)
		if err != nil {
			return nil, err
		}
	}
	return resp, err
}

// Parse 'tx_json' (which has been returned as raw json) into Transaction.Jsondetails
func decodeTxJson(resp *GetTransactionPoolResponse) error {
	for i, tx := range resp.Transactions {
		txJson := tx.TxJson
		if txJson == "" {
			return fmt.Errorf("json decode error: no TxJson info for tx #%d", i)
		}
		r := bytes.NewBufferString(txJson)
		dec := json.NewDecoder(r)
		err := dec.Decode(&tx.JsonDetails)
		if err != nil {
			return err
		}
		resp.Transactions[i].JsonDetails = tx.JsonDetails
	}
	return nil
}
