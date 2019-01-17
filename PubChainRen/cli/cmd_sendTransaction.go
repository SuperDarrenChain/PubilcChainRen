package cli

import (
	"PubChainRen/block"
	"PubChainRen/wallet"
	"fmt"

	"os"
)

/**
 *  添加区块方法
 */
func (cli *CommandLine) SendTransaction(fromData []string, toData []string, amount []int) {
	if !block.IsChainExit(cli.NODE_ID) {
		fmt.Println("区块链还未创建，请先创建区块链")
		CommandUsage()
		return
	}
	if (len(fromData) != len(toData) || len(amount) != len(amount)) {
		fmt.Println("转账参数不匹配，请确保from，to，amount参数序列正确匹配")
		CommandUsage()
		return
	}
	//判断地址是否合法
	length := len(fromData)

	for i := 0; i < length; i++ {
		if !wallet.IsAddressValid(fromData[i]) {
			fmt.Println("转账发起人地址输入无效,请重新输入")
			os.Exit(1)
		}
		if !wallet.IsAddressValid(toData[i]) {
			fmt.Println("转账接收者地址输入无效,请重新输入")
			os.Exit(1)
		}
	}
	blockChain := block.GetBlockChain(cli.NODE_ID)
	blockChain.SendTransactions(fromData, toData, amount,cli.NODE_ID)

	//更新utxoset
	utxoSet := &block.UTXOSet{BlockChain: blockChain}
	utxoSet.Update()

	//关闭数据库
	blockChain.DB.Close()

}
