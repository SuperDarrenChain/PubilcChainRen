package p2p

import (
	"PubChainRen/block"
	"PubChainRen/utils"
	"fmt"
	"net"
	"io/ioutil"

)

const (
	COMMAND_VERSION    = "version"
	COMMAND_GETBLOCKS  = "getblocks"
	COMMAND_SENDBLOCKS = "sendblocks"
)

var nodeAddress string
var chain *block.BlockChain

/**
 * 启动节点服务器
 */
func StartServer(node_id string) {
	nodeAddress = fmt.Sprintf("127.0.0.1:%s", node_id)

	/**
	 * 给全节点发送版本信息
	 */
	if nodeAddress != "127.0.0.1:3000" {
		SendVersion("127.0.0.1:3000", 0, node_id)
	}

	//监听端口
	listen, err := net.Listen("tcp", nodeAddress)
	if err != nil {
		panic(err.Error())
	}
	defer listen.Close()

	for {

		conn, err := listen.Accept()
		if err != nil {
			panic(err.Error())
		}
		//启动一个协程
		go HandleConnection(conn, node_id)
	}
}

/**
 * 处理conn连接
 */
func HandleConnection(conn net.Conn, node_id string) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		panic(err.Error())
	}

	command := utils.BytesToCommand(request[:utils.COMMANDLENGTH])

	if chain == nil {
		chain = block.GetBlockChain(node_id)
	}

	switch command {
	case COMMAND_VERSION: //版本处理
		HandleVersion(chain, request, node_id)
	case COMMAND_GETBLOCKS: //获取全节点区块信息
		HandleGetBlocks(chain, request, node_id)
	case COMMAND_SENDBLOCKS:
		//其他节点处理全节点返回的区块信息
		HandleSendblocks(request)
	default:
		fmt.Println(" 没有定义的命令 ")
	}
}
