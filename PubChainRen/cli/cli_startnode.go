package cli

import (
	"PubChainRen/p2p"
)

/**
 * 启动节点程序
 */
func (cli *CommandLine) StartNode() {

	//启动节点服务器
	p2p.StartServer(cli.NODE_ID)
}
