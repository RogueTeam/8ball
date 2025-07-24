// Inspired by https://github.com/moneroexamples/private-testnet
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal" // For signal handling
	"syscall"   // For specific signals like SIGINT, SIGTERM
)

const (
	baseDir             = "./monero-wallets"  // Changed base directory for wallet data
	walletRpcExecutable = "monero-wallet-rpc" // Ensure this is in your PATH or provide full path
	defaultTestnetPort  = "28081"             // Default Monero testnet RPC port
)

func main() {
	// Define and parse the command-line flags
	walletRpcPort := flag.String("rpc-port", "22222", "Port for the monero-wallet-rpc to listen on.")
	daemonAddress := flag.String("daemon-address", "127.0.0.1:"+defaultTestnetPort, "Address of the monerod daemon to connect to.")
	flag.Parse()

	// Prepare directories
	fmt.Printf("Creating base directory %s...\n", baseDir)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		log.Fatalf("Error creating base directory %s: %v", baseDir, err)
	}

	fmt.Printf("\nStarting Monero Wallet RPC on port %s, connecting to daemon at %s...\n", *walletRpcPort, *daemonAddress)

	walletRpcCmd := exec.Command(walletRpcExecutable,
		"--testnet", // Connect to testnet
		"--disable-rpc-ban",
		"--trusted-daemon",
		"--non-interactive",
		"--rpc-bind-ip", "127.0.0.1",
		"--rpc-bind-port", *walletRpcPort,
		"--rpc-login", "username:password",
		"--daemon-address", *daemonAddress,
		"--wallet-dir", baseDir,
		"--log-level", "0",
	)
	walletRpcCmd.Stdout = os.Stdout
	walletRpcCmd.Stderr = os.Stderr

	err := walletRpcCmd.Start()
	if err != nil {
		log.Fatalf("Error starting monero-wallet-rpc: %v", err)
	}
	fmt.Printf("Monero Wallet RPC started on 127.0.0.1:%s (PID: %d). Logs written to stdout and stderr\n", *walletRpcPort, walletRpcCmd.Process.Pid)

	// Defer stopping the wallet RPC process
	defer func() {
		if walletRpcCmd != nil && walletRpcCmd.Process != nil {
			fmt.Printf("Stopping Monero Wallet RPC (PID: %d)...\n", walletRpcCmd.Process.Pid)
			if err := walletRpcCmd.Process.Kill(); err != nil {
				log.Printf("Error killing wallet RPC process: %v", err)
			}
			_ = walletRpcCmd.Wait()
			fmt.Println("Monero Wallet RPC stopped.")
		}
	}()

	fmt.Println("\nMonero Wallet RPC is running.")
	fmt.Println("Check the log file in the 'wallet_rpc_data' directory for status.")
	fmt.Println("Press Ctrl+C to stop the process.")

	// Keep the main process running indefinitely until an interrupt signal is received.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan // Blocks until a signal is received
	fmt.Println("\nReceived termination signal. Shutting down...")
}
