package main

// This runs against the regtest harness.sh -- adapt for your needs

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"anarchy.ttfm.onion/gateway/blockchains/monero/walletrpc/old_rpc"
	"anarchy.ttfm.onion/gateway/blockchains/monero/walletrpc/rpc"
)

/////////////////////////
// Run Regtest Harness //
/////////////////////////

func main() {
	ctx := context.Background()

	charlie := rpc.New(rpc.Config{
		// charlie address
		Address: "http://127.0.0.1:28284/json_rpc",
		Client:  &http.Client{ /*default no auth HTTP client*/ },
	})

	// charlie sends 0.000333 XMR -> fred
	tr_req := rpc.TransferRequest{
		Destinations: []rpc.Destination{
			{
				Amount:  333000000,
				Address: "494aSG3QY1C4PJf7YyDjFc9n2uAuARWSoGQ3hrgPWXtEjgGrYDn2iUw8WJP5Dzm4GuMkY332N9WfbaKfu5tWM3wk8ZeSEC5",
			},
		},
		AccountIndex:   0,
		SubaddrIndices: []uint64{0},
		Priority:       0,
		RingSize:       11,
		UnlockTime:     20, // locked tx .. maybe for atomic swap
		GetTxKey:       true,
	}
	tr_resp, err := charlie.Transfer(ctx, &tr_req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("tx_hash:", tr_resp.TxHash)

	daemon := old_rpc.New(old_rpc.Config{
		Address: "http://127.0.0.1:18081",
	})

	// get daemon txs info
	gtxs_req := old_rpc.GetTransactionsRequest{
		TxsHashes: []string{
			tr_resp.TxHash,
		},
		DecodeAsJson: true,
	}

	gtxs_resp, err := daemon.GetTransactions(ctx, &gtxs_req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("status:", gtxs_resp.Status)

	// try resend tx again direct to daemon - Not Relayed
	srt_req := old_rpc.SendRawTransactionRequest{
		TxAsHex:    gtxs_resp.TxsAsHex[0],
		DoNotRelay: false,
	}

	srt_resp, err := daemon.SendRawTransaction(ctx, &srt_req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("status:", srt_resp.Status)

	// get mempool
	gtp_resp, err := daemon.GetTransactionPool(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("status:", gtp_resp.Status)

	client := rpc.New(rpc.Config{
		Address: "http://127.0.0.1:18081/json_rpc",
		Client:  &http.Client{ /*default no auth HTTP client*/ },
	})

	// mine to fred
	drg_req := rpc.DemonRegtestGenerateRequest{
		AmountOfBlocks: 1,
		WalletAddress:  "494aSG3QY1C4PJf7YyDjFc9n2uAuARWSoGQ3hrgPWXtEjgGrYDn2iUw8WJP5Dzm4GuMkY332N9WfbaKfu5tWM3wk8ZeSEC5",
		StartingNonce:  0,
	}

	drg_resp, err := client.DaemonRegtestGenerate(ctx, &drg_req)
	if err != nil {
		// if not regtest daemon will say:
		// "json2.Error {Code: -13, Message: "Regtest required when generating blocks"}"
		log.Fatal(err)
	}
	fmt.Println("status:", drg_resp.Status)

	//  get daemon height,etc.
	dgi_resp, err := client.DaemonGetInfo(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("status:", dgi_resp.Status)
}
