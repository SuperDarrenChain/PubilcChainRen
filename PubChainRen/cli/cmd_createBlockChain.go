package cli

import (
	"PubChainRen/block"
	"PubChainRen/transaction"
	"PubChainRen/wallet"
	"fmt"
	"os"
)

/**
 * 创建区块链，指定自己的data
 */
func (cli *CommandLine) CreateBlockChain(address string) {

	//1.首先判断地址是否有效
	if !wallet.IsAddressValid(address) {
		fmt.Println("地址输入无效,请重新输入")
		os.Exit(1)
	}
	//2.创建一个coinbase交易
	txCoinbase := transaction.NewCoinbaseTransaction(address)
	block.CreateBlockChainWithGenesisBlock([]*transaction.Transaction{txCoinbase},cli.NODE_ID)
}
