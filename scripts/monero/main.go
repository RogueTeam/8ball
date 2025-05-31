// Inspired by https://github.com/moneroexamples/private-testnet
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal" // For signal handling
	"path/filepath"
	"syscall" // For specific signals like SIGINT, SIGTERM
)

const (
	baseDir             = "./nodes_data"
	nodeBaseName        = "node"
	monerodExecutable   = "monerod"           // Ensure this is in your PATH or provide full path
	walletRpcExecutable = "monero-wallet-rpc" // Ensure this is in your PATH or provide full path
)

// NodeConfig holds the configuration for each Monero node
type NodeConfig struct {
	ID             int
	P2PPort        int
	RPCPort        int
	ZMQRPCPort     int
	DataDir        string
	ExclusiveNodes []string
	LogFile        string
	MiningAddress  string // Added to store the mining address for this specific node
}

func main() {
	// 1. Define and parse the command-line flags
	miningWalletAddress := flag.String("mining-wallet-address", "", "A valid Monero testnet address for the first node to mine to.")
	walletRpcEnabled := flag.Bool("wallet-rpc", false, "Set to true to start monero-wallet-rpc.")
	flag.Parse()

	// Prepare directories
	fmt.Printf("Creating base directory %s...\n", baseDir)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		log.Fatalf("Error creating base directory %s: %v", baseDir, err)
	}

	// Define the configurations for the three nodes
	nodeConfigs := []NodeConfig{
		{
			ID:         1,
			P2PPort:    28080,
			RPCPort:    28081, // Default RPC port for node 1
			ZMQRPCPort: 28082, // Default ZMQ RPC port for node 1
		},
		{
			ID:         2,
			P2PPort:    38080,
			RPCPort:    38081,
			ZMQRPCPort: 38082,
		},
		{
			ID:         3,
			P2PPort:    48080,
			RPCPort:    48081,
			ZMQRPCPort: 48082,
		},
	}

	// Populate data and log paths for each config
	for i := range nodeConfigs {
		nodeConfigs[i].DataDir = filepath.Join(baseDir, fmt.Sprintf("%s_%02d", nodeBaseName, nodeConfigs[i].ID))
		nodeConfigs[i].LogFile = filepath.Join(baseDir, fmt.Sprintf("%s_%02d.log", nodeBaseName, nodeConfigs[i].ID))

		// Set the mining address for the first node if the flag is provided
		if nodeConfigs[i].ID == 1 && *miningWalletAddress != "" {
			nodeConfigs[i].MiningAddress = *miningWalletAddress
		}
	}

	// Set exclusive nodes for each config (interconnect all)
	for i := range nodeConfigs {
		for j := range nodeConfigs {
			if i != j {
				nodeConfigs[i].ExclusiveNodes = append(nodeConfigs[i].ExclusiveNodes, fmt.Sprintf("127.0.0.1:%d", nodeConfigs[j].P2PPort))
			}
		}
	}

	// Start Monero nodes
	fmt.Println("\nStarting Monero testnet nodes...")
	nodeProcesses := make(map[int]*exec.Cmd)
	for _, config := range nodeConfigs {
		cmd, err := startMonerod(config)
		if err != nil {
			log.Fatalf("Error starting monerod for node %d: %v", config.ID, err)
		}
		nodeProcesses[config.ID] = cmd
		fmt.Printf("Monero node %d started (PID: %d). Logs written to %s\n", config.ID, cmd.Process.Pid, config.LogFile)

		// Inform about mining status
		if config.ID == 1 && config.MiningAddress != "" {
			fmt.Printf("Node 1 is configured to mine to address: %s\n", config.MiningAddress)
		}
	}

	// ---
	// Defer stopping node processes
	defer func() {
		fmt.Println("\nStopping Monero nodes...")
		for id, cmd := range nodeProcesses {
			if cmd != nil && cmd.Process != nil {
				fmt.Printf("Stopping node %d (PID: %d)\n", id, cmd.Process.Pid)
				if err := cmd.Process.Kill(); err != nil {
					log.Printf("Error killing process for node %d: %v", id, err)
				}
				// Wait for the process to exit to clean up resources
				_ = cmd.Wait()
			}
		}
		fmt.Printf("All nodes stopped. Data is in %s\n", baseDir)
	}()

	var walletRpcCmd *exec.Cmd
	const walletRpcPort = "22222" // Define the RPC port for the wallet-rpc

	// Conditionally start monero-wallet-rpc only if the 'wallet-rpc' flag is set to true
	if *walletRpcEnabled {
		fmt.Printf("\nStarting Monero Wallet RPC on port %s...\n", walletRpcPort)
		walletDir := filepath.Join(baseDir, "wallets_rpc") // Create a dedicated directory for wallet RPC data
		if err := os.MkdirAll(walletDir, 0755); err != nil {
			log.Fatalf("Error creating wallet RPC directory %s: %v", walletDir, err)
		}

		walletLogFile, err := os.Create(filepath.Join(walletDir, "wallet_rpc.log"))
		if err != nil {
			log.Fatal(err)
		}

		walletRpcCmd = exec.Command(walletRpcExecutable,
			"--testnet",
			"--disable-rpc-ban",
			"--trusted-daemon",
			"--non-interactive",
			"--rpc-bind-ip", "127.0.0.1",
			"--rpc-bind-port", walletRpcPort, // Use the defined port
			"--daemon-address", "127.0.0.1:28081", // Connect to the first node's RPC port
			"--rpc-login", "username:password", // Replace with actual username/password for security
			"--wallet-dir", walletDir,
			"--log-file", walletLogFile.Name(),
		)
		walletRpcCmd.Stdout = walletLogFile
		walletRpcCmd.Stderr = walletLogFile

		err = walletRpcCmd.Start()
		if err != nil {
			log.Fatalf("Error starting monero-wallet-rpc: %v", err)
		}
		fmt.Printf("Monero Wallet RPC started on 127.0.0.1:%s (PID: %d). Logs written to %s\n", walletRpcPort, walletRpcCmd.Process.Pid, walletLogFile.Name())

		// Defer stopping the wallet RPC process
		defer func() {
			if walletRpcCmd != nil && walletRpcCmd.Process != nil {
				fmt.Printf("Stopping Monero Wallet RPC (PID: %d)...\n", walletRpcCmd.Process.Pid)
				if err := walletRpcCmd.Process.Kill(); err != nil {
					log.Printf("Error killing wallet RPC process: %v", err)
				}
				_ = walletRpcCmd.Wait()
			}
		}()
	} else {
		fmt.Println("\n'wallet-rpc' flag not set. Monero Wallet RPC will not be started.")
	}

	fmt.Println("\nPrivate Monero testnet nodes are running and attempting to synchronize.")
	fmt.Println("Check the log files in the 'nodes_data' directory for status.")
	fmt.Println("Press Ctrl+C to stop all processes.")

	// Keep the main process running indefinitely until an interrupt signal is received.
	// This allows the child processes (monerod and monero-wallet-rpc) to run in the background.
	// When Ctrl+C is pressed, the deferred functions will handle cleanup.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan // Blocks until a signal is received
	fmt.Println("\nReceived termination signal. Shutting down...")
}

// startMonerod starts a Monero daemon process
func startMonerod(config NodeConfig) (*exec.Cmd, error) {
	nodeDir := config.DataDir
	if err := os.MkdirAll(nodeDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating node directory %s: %v", nodeDir, err)
	}

	cmdArgs := []string{
		"--testnet",
		"--no-igd",
		"--hide-my-port",
		"--data-dir", nodeDir,
		"--p2p-bind-ip", "127.0.0.1",
		"--non-interactive", // Prevents monerod from waiting for user input
		"--log-level", "0",
		"--fixed-difficulty", "100", // Keep this for easy block generation on testnet
		"--disable-rpc-ban",
		"--p2p-bind-port", fmt.Sprintf("%d", config.P2PPort),
		"--rpc-bind-port", fmt.Sprintf("%d", config.RPCPort),
		"--zmq-rpc-bind-port", fmt.Sprintf("%d", config.ZMQRPCPort),
	}

	for _, node := range config.ExclusiveNodes {
		cmdArgs = append(cmdArgs, "--add-exclusive-node", node)
	}

	// Conditionally add mining arguments if a mining address is provided for this node
	if config.MiningAddress != "" {
		cmdArgs = append(cmdArgs, "--start-mining", config.MiningAddress)
		cmdArgs = append(cmdArgs, "--mining-threads", "1") // You can adjust the number of threads
		log.Printf("Node %d will attempt to mine to address: %s", config.ID, config.MiningAddress)
	}

	logFile, err := os.Create(config.LogFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file for node %d: %w", config.ID, err)
	}

	cmd := exec.Command(monerodExecutable, cmdArgs...)
	cmd.Stdout = logFile // Redirect stdout to log file
	cmd.Stderr = logFile // Redirect stderr to log file

	if err := cmd.Start(); err != nil {
		logFile.Close() // Close on error during startup
		return nil, fmt.Errorf("failed to start monerod for node %d: %w", config.ID, err)
	}

	return cmd, nil
}
