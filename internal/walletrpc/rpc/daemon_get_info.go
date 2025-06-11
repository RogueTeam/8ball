package rpc

import "context"

type DemonGetInfoResponse struct {
	Status                      string `json:"status"`
	AdjustedTime                uint64 `json:"adjusted_time"`
	AltBlocksCount              uint64 `json:"alt_blocks_count"`
	BlockweightLimit            uint64 `json:"block_weight_limit"`
	BlockWeightMedian           uint64 `json:"block_weight_median"`
	BootstrapDaemonAddress      string `json:"bootstrap_daemon_address"`
	BusySyncing                 bool   `json:"busy_syncing"`
	Credits                     uint64 `json:"credits"`
	CumulaticeDifficulty        uint64 `json:"cumulative_difficulty"`
	CumulaticeDifficultyTop64   uint64 `json:"cumulative_difficulty_top64"`
	DatabasSize                 uint64 `json:"database_size"`
	Difficulty                  uint64 `json:"difficulty"`
	DifficultyTop64             uint64 `json:"difficulty_top64"`
	FreeSpace                   uint64 `json:"free_space"`
	GreyPeerlistSize            uint64 `json:"grey_peerlist_size"`
	Height                      uint64 `json:"height"`
	HeightWithoutBootstrap      uint64 `json:"height_without_bootstrap"`
	IncomingConnectionsCount    uint64 `json:"incoming_connections_count"`
	Mainnet                     bool   `json:"mainnet"`
	NetType                     string `json:"nettype"`
	Offline                     bool   `json:"offline"`
	OutgoingConnectionsCount    uint64 `json:"outgoing_connections_count"`
	Restricted                  bool   `json:"restricted"`
	RpcConnectionsCount         uint64 `json:"rpc_connections_count"`
	Stagenet                    bool   `json:"stagenet"`
	StartTime                   uint64 `json:"start_time"`
	Sychronized                 bool   `json:"synchronized"`
	Target                      uint64 `json:"target"`
	TargetHeight                uint64 `json:"target_height"`
	Testnet                     bool   `json:"testnet"`
	TopBlockHash                string `json:"top_block_hash"`
	TopHash                     string `json:"top_hash"`
	TxCount                     uint64 `json:"tx_count"`
	TxPoolSize                  uint64 `json:"tx_pool_size"`
	Untrusted                   bool   `json:"untrusted"`
	UpdateAvailable             bool   `json:"update_available"`
	Version                     string `json:"version"`
	WasBootstrapEverUsed        bool   `json:"was_bootstrap_ever_used"`
	WhitePeerlistSize           uint64 `json:"white_peerlist_size"`
	WideCumulativeDifficultyHex string `json:"wide_cumulative_difficulty"`
	WideDifficultyHex           string `json:"wide_difficulty"`
}

// Return all basic info from the daemon
func (c *Client) DaemonGetInfo(ctx context.Context) (*DemonGetInfoResponse, error) {
	resp := &DemonGetInfoResponse{}
	err := c.Do(ctx, "get_info", nil, resp)
	return resp, err
}
