package utils

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/hyperdrive-stakewise/adapter/utils/terminal"
	swclient "github.com/nodeset-org/hyperdrive-stakewise/client"

	"github.com/rocket-pool/node-manager-core/wallet"
)

func CheckIfWalletReady(hd *swclient.HyperdriveClient) (wallet.WalletStatus, bool, error) {
	// Get & check wallet status
	statusResponse, err := hd.Api.Wallet.Status()
	if err != nil {
		return wallet.WalletStatus{}, false, err
	}
	status := statusResponse.Data.WalletStatus

	// Check if it's already set properly and the wallet has been loaded
	if !wallet.IsWalletReady(status) {
		fmt.Println("The node wallet is not loaded or your node is in read-only mode. Please run `hyperdrive wallet status` for more details.")
		return status, false, nil
	}
	return status, true, nil
}

// Verifies the daemon has a node address ready and loaded (allows for masquerade mode support).
func CheckIfAddressReady(hd *swclient.HyperdriveClient) (wallet.WalletStatus, bool, error) {
	// Get & check wallet status
	statusResponse, err := hd.Api.Wallet.Status()
	if err != nil {
		return wallet.WalletStatus{}, false, err
	}
	status := statusResponse.Data.WalletStatus

	// There's an address ready
	if status.Address.HasAddress {
		if status.Address.NodeAddress != status.Wallet.WalletAddress {
			fmt.Printf("%sReminder: You are currently masquerading as %s%s%s.\nYou can create transactions but cannot sign or submit them.%s\n", terminal.ColorGreen, terminal.ColorBlue, status.Address.NodeAddress, terminal.ColorGreen, terminal.ColorReset)
			fmt.Println()
		}
		return status, true, nil
	}

	// If the address isn't ready, check if the wallet's ready
	if !status.Wallet.IsLoaded {
		if !status.Wallet.IsOnDisk {
			fmt.Println("The node wallet has not been initialized yet. Please run `hyperdrive wallet init` or `hyperdrive wallet recover` first, then run this again.")
			return status, false, nil
		}
		fmt.Println("The daemon requires your node wallet's password to unlock it. Please run `hyperdrive wallet set-password` first, then run this again.")
		return status, false, nil
	}

	// The address isn't ready but the wallet is so have the user run restore-address to fix it
	fmt.Printf("The node wallet is %s%s%s but the node address is not set. Please restore it with `hyperdrive wallet restore-address` or `hyperdrive wallet masquerade` first, then run this again.", terminal.ColorBlue, status.Wallet.WalletAddress, terminal.ColorReset)
	return status, false, nil
}

// Print a TX's details to the console.
func PrintTransactionHash(hd *swclient.HyperdriveClient, hash common.Hash) {
	finalMessage := "Waiting for the transaction to be included in a block... you may wait here for it, or press CTRL+C to exit and return to the terminal.\n\n"
	printTransactionHashImpl(hd, hash, finalMessage)
}

// Print a TX's details to the console, but inform the user NOT to cancel it.
func PrintTransactionHashNoCancel(hd *swclient.HyperdriveClient, hash common.Hash) {
	finalMessage := "Waiting for the transaction to be included in a block... **DO NOT EXIT!** This transaction is one of several that must be completed.\n\n"
	printTransactionHashImpl(hd, hash, finalMessage)
}

// Print a batch of transaction hashes to the console.
func PrintTransactionBatchHashes(hd *swclient.HyperdriveClient, hashes []common.Hash) {
	finalMessage := "Waiting for the transactions to be included in one or more blocks... you may wait here for them, or press CTRL+C to exit and return to the terminal.\n\n"

	// Print the hashes
	fmt.Println("Transactions have been submitted with the following hashes:")
	hashStrings := make([]string, len(hashes))
	for i, hash := range hashes {
		hashString := hash.String()
		hashStrings[i] = hashString
		fmt.Println(hashString)
	}
	fmt.Println()

	txWatchUrl := getTxWatchUrl(hd)
	if txWatchUrl != "" {
		fmt.Println("You may follow their progress by visiting the following URLs in sequence:")
		for _, hash := range hashStrings {
			fmt.Printf("%s/%s\n", txWatchUrl, hash)
		}
	}
	fmt.Println()

	fmt.Print(finalMessage)
}

// Print a warning to the console if the user set a custom nonce, but this operation involves multiple transactions
func PrintMultiTransactionNonceWarning() {
	fmt.Printf("%sNOTE: You have specified the `nonce` flag to indicate a custom nonce for this transaction.\n"+
		"However, this operation requires multiple transactions.\n"+
		"Hyperdrive will use your custom value as a basis, and increment it for each additional transaction.\n"+
		"If you have multiple pending transactions, this MAY OVERRIDE more than the one that you specified.%s\n\n", terminal.ColorYellow, terminal.ColorReset)
}

// Implementation of PrintTransactionHash and PrintTransactionHashNoCancel
func printTransactionHashImpl(hd *swclient.HyperdriveClient, hash common.Hash, finalMessage string) {
	txWatchUrl := getTxWatchUrl(hd)
	hashString := hash.String()
	fmt.Printf("Transaction has been submitted with hash %s.\n", hashString)
	if txWatchUrl != "" {
		fmt.Printf("You may follow its progress by visiting:\n")
		fmt.Printf("%s/%s\n\n", txWatchUrl, hashString)
	}
	fmt.Print(finalMessage)
}

// Get the URL for watching the transaction in a block explorer
func getTxWatchUrl(hd *swclient.HyperdriveClient) string {
	cfg, isNew, err := hd.LoadConfig()
	if err != nil {
		fmt.Printf("Warning: couldn't read config file so the transaction URL will be unavailable (%s).\n", err)
		return ""
	}

	if isNew {
		fmt.Print("Settings file not found. Please run `hyperdrive service config` to set up Hyperdrive.")
		return ""
	}
	return cfg.HyperdriveResources.TxWatchUrl
}
