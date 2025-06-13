package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// Define command-line flags.
	// Default values are empty strings, meaning if the flag isn't provided,
	// monerod will use its own internal defaults for data directory and network.
	dataDirFlag := flag.String("data-dir", "", "Directory for Monero blockchain data (defaults to monerod's default if not specified)")
	networkFlag := flag.String("network", "", "Monero network to use (mainnet, testnet, stagenet). Defaults to mainnet if not specified.")

	flag.Parse() // Parse the command-line flags

	// Define the path to the monerod executable.
	// It's assumed to be in your system's PATH. If not, replace "monerod"
	// with the full path, e.g., "/usr/local/bin/monerod" or "C:\\Monero\\monerod.exe"
	monerodPath := "monerod"

	// Define the base arguments for the monerod command.
	args := []string{
		"--prune-blockchain",
		// "--proxy", "127.0.0.1:9050",
		// "--tx-proxy", "tor,127.0.0.1:9050,10",
		"--p2p-bind-ip", "127.0.0.1",
		"--no-igd",
		"--non-interactive",
		// "--detach",
		"--log-level", "0",
	}

	// Conditionally add --data-dir if the flag was provided
	var actualDataDir string
	if *dataDirFlag != "" {
		// Resolve the absolute path for the data directory if provided
		resolvedDataDir, err := filepath.Abs(*dataDirFlag)
		if err != nil {
			log.Fatalf("Error resolving absolute path for data directory %s: %v", *dataDirFlag, err)
		}
		actualDataDir = resolvedDataDir
		// Ensure the specified data directory exists
		if err := os.MkdirAll(actualDataDir, 0755); err != nil {
			log.Fatalf("Error creating data directory %s: %v", actualDataDir, err)
		}
		args = append(args, "--data-dir", actualDataDir)
	} else {
		// If --data-dir is not provided, monerod will use its default.
		// We'll try to determine what that default would be for informational purposes.
		if os.Getenv("HOME") != "" { // Linux/macOS
			actualDataDir = filepath.Join(os.Getenv("HOME"), ".monero", "lmdb") // monerod's default is usually lmdb inside .monero
		} else if os.Getenv("APPDATA") != "" { // Windows
			actualDataDir = filepath.Join(os.Getenv("APPDATA"), "Monero", "lmdb")
		} else {
			actualDataDir = "monerod's default (likely current directory or system-specific)" // Best guess
		}
	}

	// Conditionally add network-specific arguments based on the --network flag
	var actualNetwork string
	switch *networkFlag {
	case "mainnet", "": // If --network is "mainnet" or not provided, monerod uses mainnet by default
		actualNetwork = "mainnet"
		// No additional argument needed for mainnet
	case "testnet":
		actualNetwork = "testnet"
		args = append(args, "--testnet")
	case "stagenet":
		actualNetwork = "stagenet"
		args = append(args, "--stagenet")
	default:
		log.Fatalf("Invalid network specified: %s. Use 'mainnet', 'testnet', or 'stagenet'.", *networkFlag)
	}

	fmt.Printf("Starting monerod on %s network.\n", actualNetwork)
	if *dataDirFlag != "" {
		fmt.Printf("Using specified data directory: %s\n", actualDataDir)
	} else {
		fmt.Printf("Using monerod's default data directory (usually %s).\n", actualDataDir)
	}
	fmt.Printf("Full command: %s %v\n", monerodPath, args)

	// Create a new command.
	cmd := exec.Command(monerodPath, args...)

	// Redirect standard output and standard error of the monerod process
	// to the standard output and standard error of this Go program.
	// This allows you to see monerod's logs in your Go program's console.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Start the command in a new process.
	err := cmd.Start()
	if err != nil {
		log.Fatalf("Failed to start monerod: %v", err)
	}

	fmt.Printf("monerod started successfully with PID: %d\n", cmd.Process.Pid)
	fmt.Println("You can check its status by looking at the logs in your data directory")
	fmt.Println("or by looking for the process with `ps aux | grep monerod` (Linux/macOS)")
	fmt.Println("or `tasklist | findstr monerod.exe` (Windows).")

	// Wait for the command to finish.
	err = cmd.Wait()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			log.Printf("monerod exited with error: %v, Exit Status: %d", err, exitError.ExitCode())
		} else {
			log.Printf("monerod exited with error: %v", err)
		}
	} else {
		fmt.Println("monerod process finished.")
	}

	fmt.Println("Note: If monerod detached successfully, this Go program will exit,")
	fmt.Println("but monerod will continue running in the background.")
	fmt.Println("Make sure Tor is running on 127.0.0.1:9050 before executing this program.")
}
