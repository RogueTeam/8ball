package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var config struct {
	dataDir       string
	network       string
	miningAddress string
	miningThreads int
	peer          string
	useTor        bool
	interactive   bool
}

const monerod = "monerod"

func init() {
	flagset := flag.NewFlagSet("monerod", flag.ExitOnError)

	flagset.BoolVar(&config.useTor, "tor", false, "Use tor")
	flagset.BoolVar(&config.interactive, "interactive", false, "Interactive mode")
	flagset.StringVar(&config.dataDir, "data-dir", "", "Directory for Monero blockchain data (defaults to monerod's default if not specified)")
	flagset.StringVar(&config.network, "network", "", "Monero network to use (mainnet, testnet, stagenet). Defaults to mainnet if not specified.")
	flagset.StringVar(&config.miningAddress, "mining-address", "", "Address for mining.")
	flagset.IntVar(&config.miningThreads, "mining-threads", 1, "Threads for mining.")
	flagset.StringVar(&config.peer, "peer", "", "Peer")

	err := flagset.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
}

func prepareArgs() (args []string) {
	args = []string{
		// "--prune-blockchain",
		"--p2p-bind-ip", "127.0.0.1",
		"--no-igd",
		// "--detach",
		"--log-level", "0",
	}

	// Conditionally add network-specific arguments based on the --network flag
	var actualNetwork string
	switch config.network {
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
		log.Fatalf("Invalid network specified: %s. Use 'mainnet', 'testnet', or 'stagenet'.", config.network)
	}

	if !config.interactive {
		args = append(args, "--non-interactive")
	}

	if config.useTor {
		args = append(args, "--proxy", "127.0.0.1:9050")
		if strings.Contains(config.peer, ".onion") {
			args = append(args, "--tx-proxy", "tor,127.0.0.1:9050,10")
		}
	}

	var actualDataDir string
	if config.dataDir != "" {
		// Resolve the absolute path for the data directory if provided
		resolvedDataDir, err := filepath.Abs(config.dataDir)
		if err != nil {
			log.Fatalf("Error resolving absolute path for data directory %s: %v", config.dataDir, err)
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
		if runtime.GOOS == "linux" || runtime.GOOS == "darwin" || runtime.GOOS == "freebsd" { // Linux/macOS
			actualDataDir = filepath.Join(os.Getenv("HOME"), ".bitmonero") // monerod's default is usually lmdb inside .monero
			if config.network != "mainnet" {
				actualDataDir = filepath.Join(actualDataDir, config.network)
			}
		} else if runtime.GOOS == "windows" { // Windows
			actualDataDir = "C:/ProgramData/bitmonero"
			if config.network != "mainnet" {
				actualDataDir = filepath.Join(actualDataDir, config.network)
			}
		} else {
			actualDataDir = "monerod's default (likely current directory or system-specific)" // Best guess
		}
	}

	if config.miningAddress != "" {
		log.Println("Mining to address", config.miningAddress)
		args = append(args, "--start-mining", config.miningAddress)
		args = append(args, "--mining-threads", strconv.Itoa(config.miningThreads))
	}

	if config.peer != "" {
		log.Println("Adding peer", config.peer)
		args = append(args, "--add-peer", config.peer)
		args = append(args, "--add-priority-node", config.peer)
	}

	fmt.Printf("Starting monerod on %s network.\n", actualNetwork)
	if config.dataDir != "" {
		fmt.Printf("Using specified data directory: %s\n", actualDataDir)
	} else {
		fmt.Printf("Using monerod's default data directory (usually %s).\n", actualDataDir)
	}
	fmt.Printf("Full command: %s %v\n", monerod, strings.Join(args, " "))
	return args
}

func main() {
	cmd := exec.Command(monerod, prepareArgs()...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Start()
	if err != nil {
		log.Fatalf("Failed to start monerod: %v", err)
	}

	fmt.Printf("monerod started successfully with PID: %d\n", cmd.Process.Pid)
	fmt.Println("You can check its status by looking at the logs in your data directory")
	fmt.Println("or by looking for the process with `ps aux | grep monerod` (Linux/macOS)")
	fmt.Println("or `tasklist | findstr monerod.exe` (Windows).")

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
