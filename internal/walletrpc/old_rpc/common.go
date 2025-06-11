package old_rpc

// Common structs for Daemon RPC's

// Below are the structures to decode:
// 'tx_json' - GetTransactionPool
// 'as_json' - GetTransactions

type OutPubKey string

type EcdhInfo struct {
	// Diffie-Hellman elliptic curves' structures
	Amount string `json:"amount"`
	Mask   string `json:"mask"`
}

type RctSignatures struct {
	// List of signatures used in the ring of signatures
	// - to hide the true origin of the transaction inputs.
	Type   uint64      `json:"type"`
	TxnFee uint64      `json:"txnFee"`
	Ecdh   []EcdhInfo  `json:"ecdhInfo"`
	OutPk  []OutPubKey `json:"outPk"`
}

type TaggedKey struct {
	StealthKey string `json:"key"`
	ViewTag    string `json:"view_tag"`
}

type Target struct {
	TaggedKey TaggedKey `json:"tagged_key"`
}

type Output struct {
	AmountCoinbase uint64 `json:"amount"`
	Target         Target `json:"target"`
}

type PubKeyPrevOut struct {
	Amount     uint64   `json:"amount"`
	KeyOffsets []uint64 `json:"key_offsets"`
	KeyImage   string   `json:"k_image"`
}

type Input struct {
	Key PubKeyPrevOut `json:"key"`
}

type JsonDetails struct {
	Version       uint64        `json:"version"`
	TxUnlockTime  uint64        `json:"unlock_time"` // unlock time for txs we lock for n blocks
	Vin           []Input       `json:"vin"`
	Vout          []Output      `json:"vout"`
	Extra         []uint8       `json:"extra"` // bytes: Payment ID, miner info, random bytes, etc.
	RctSignatures RctSignatures `json:"rct_signatures"`
}
