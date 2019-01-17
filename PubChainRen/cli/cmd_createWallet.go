package cli

import (
	"PubChainRen/wallet"
)

/**
 * 创建钱包实例
 */
func (cli *CommandLine) createWallet() {
	wallets1 := wallet.NewWallets(cli.NODE_ID)
	wallets1.CreateWallet()
	//保存钱包地址列表到文件中
	wallets1.SaveToFile(cli.NODE_ID)
}
