package cli

import (
	"PubChainRen/block"
	"PubChainRen/wallet"
	"fmt"

	"os"
)

/**
 * 查询某个地址的余额
 */
func (cli *CommandLine) GetBalance(address string) {
	if !block.IsChainExit(cli.NODE_ID) {
		fmt.Println("区块链还未创建，请先创建区块链")
		CommandUsage()
		return
	}
	if !wallet.IsAddressValid(address) {
		fmt.Println("地址输入无效,请重新输入")
		os.Exit(1)
	}

	//balance := blockChain.GetBalance(address)

	blockChain := block.GetBlockChain(cli.NODE_ID)
	defer blockChain.DB.Close()
	utxoSet := &block.UTXOSet{BlockChain: blockChain}
	balance := utxoSet.GetBalance(address)
	fmt.Printf("%s共有%d个Token\n", address, balance)
}
