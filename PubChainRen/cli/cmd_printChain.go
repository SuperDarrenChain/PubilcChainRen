package cli

import (
	"PubChainRen/block"
	"fmt"

)

/**
 * 打印所支持的所有的命令及说明
 */
func (cli *CommandLine) PrintChain() {
	if !block.IsChainExit(cli.NODE_ID) {
		fmt.Println("区块链还未创建，请先创建区块链")
		CommandUsage()
		return
	}
	blockChain := block.GetBlockChain(cli.NODE_ID)
	blockChain.PrintChain()
	blockChain.DB.Close()
}
