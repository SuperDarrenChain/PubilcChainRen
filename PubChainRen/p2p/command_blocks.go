package p2p

import (
	"PubChainRen/block"
	"PubChainRen/utils"
	"bytes"
	"encoding/gob"
	"fmt"

)

/**
 * getBlocks 意为：获取一下全节点有什么区块
 */
type GetBlocks struct {
	AddrFrom    string //当前节点自己的地址
	FromIndex   int64  //从第几块数据开始同步
	CountBlocks int64  //同步多少个区块
}

/**
 * 全节点处理其他节点的请求区块信息的命令
 */
func HandleGetBlocks(chain *block.BlockChain, request []byte, node_id string) {
	getBlocksBytes := request[utils.COMMANDLENGTH:]

	//进行序列化
	var getBlocks GetBlocks
	reader := bytes.NewReader(getBlocksBytes)
	decoder := gob.NewDecoder(reader)
	err := decoder.Decode(&getBlocks)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("接收到了来自%s的数据:%s\n", getBlocks.AddrFrom, string(request[:utils.COMMANDLENGTH]))

	//获取区块信息
	blocks := chain.GetBlocks(getBlocks.FromIndex, getBlocks.CountBlocks)
	if blocks != nil {
		fmt.Println(" 数据不为空")
		blocksArray := &Blocks{blocks}
		payload := append(utils.CommandToBytes(COMMAND_SENDBLOCKS), blocksArray.Searlize()...)
		//全节点发送数据给其他节点
		SendMessage(getBlocks.AddrFrom, payload)
	} else {
		fmt.Println(" 数据为空")
	}
}

/**
 * 其他节点向全节点请求有哪些区块的信息
 */
func SendGetBlocks(toAddr string, fromIndex int64, countBlocks int64) {

	getBlocks := GetBlocks{nodeAddress, fromIndex, countBlocks}

	//序列化
	searlizeByte := utils.GobEncode(getBlocks)

	// 拼接命令 + 数据
	request := append(utils.CommandToBytes(COMMAND_GETBLOCKS), searlizeByte...)

	//发送数据给全节点
	SendMessage(toAddr, request)
}
