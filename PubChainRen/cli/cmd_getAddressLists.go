package cli

import (
	"PubChainRen/wallet"
	"fmt"
)

/**
 * 打印所有钱包地址
 */
func (cli *CommandLine) PrintAddressLists() {
	wallets := wallet.NewWallets(cli.NODE_ID)
	fmt.Println("打印所有钱包地址列表：")
	for _, wallet := range wallets.Wallets {
		fmt.Println(wallet.GetAddress())
	}
}
