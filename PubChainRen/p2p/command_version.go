package p2p

import (
	"PubChainRen/block"
	"PubChainRen/utils"
	"net"
	"io"
	"bytes"
	"encoding/gob"

	"fmt"

)

/**
 * 处理版本
 */
func HandleVersion(chain *block.BlockChain, request []byte, node_id string) {

	//1.从request中获取版本的数据：[]byte
	versionBytes := request[utils.COMMANDLENGTH:]

	//反序列化
	var version Version
	reader := bytes.NewReader(versionBytes)
	decoder := gob.NewDecoder(reader)
	err := decoder.Decode(&version)
	if err != nil {
		panic(err.Error())
	}

	clientHeight := version.LastHeight

	//获取自己的blockchain和最新的区块高度
	lastHeight := chain.GetLastHeight()
	if lastHeight > clientHeight {
		//发送自己的版本给client
		fmt.Println("服务端接收到客户端节点的信息:", version.AddFrom)
		SendVersion(version.AddFrom, lastHeight, node_id)
	} else {
		//去获取数据
		fmt.Println("客户端节点接收到服务端的数据信息:", version)
		SendGetBlocks(version.AddFrom, lastHeight+1, version.LastHeight-(lastHeight+1))
	}
}

//发送version信息
func SendVersion(toAddr string, lastHeight int64, node_id string) {
	//1.获取当前区块链最新区块高度
	//chain := block.GetBlockChain(node_id)
	//bestHeight := chain.GetLastHeight()

	//2.创建version对象,第三个参数是要发的是自己节点的地址
	version := Version{utils.NODE_VERSION, lastHeight, nodeAddress}

	//3.将version序列化
	payload := utils.GobEncode(version)

	//4.拼接命令+数据
	request := append(utils.CommandToBytes(COMMAND_VERSION), payload...)

	//5.发送数据
	SendMessage(toAddr, request)
}

/**
 * 发送命令及消息
 */
func SendMessage(to string, data []byte) {
	conn, err := net.Dial("tcp", to)
	if err != nil {
		panic(err.Error())
	}
	defer conn.Close()
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		panic(err.Error())
	}
}
