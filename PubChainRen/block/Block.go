package block

import (
	"PubChainRen/merkletree"
	"PubChainRen/pow"
	"PubChainRen/transaction"
	"time"
	"bytes"
	"fmt"
	"encoding/gob"

)

/**
 * 区块结构体
 */
type Block struct {
	//1.区块高度
	Height int64
	//2.上一个区块hash
	PreHash []byte
	//3.交易数据
	//Data []byte
	Txs []*transaction.Transaction
	//4.时间戳
	TimeStamp int64
	//5.Hash
	Hash []byte

	//6.Nonce 随机值
	Nonce int64
}

func (block *Block) GetHeight() int64 {
	return block.Height
}

func (block *Block) GetPreHash() []byte {
	return block.PreHash
}

/**
 * 返回所有交易的hash的hash值
 */
func (block *Block) GetTxsHash() []byte {
	//var txHashes [][]byte
	//	//var txHash [32]byte
	//	//for _, tx := range block.Txs {
	//	//	txHashes = append(txHashes, tx.TxHash)
	//	//}
	//	//txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	//return txHash[:]

	var txs [][]byte
	for _, tx := range block.Txs {
		txs = append(txs, tx.Serialize())
	}
	mTree := merkletree.NewMerkleTree(txs)
	return mTree.RootNode.DataHash
}

func (block *Block) GetTimeStamp() int64 {
	return block.TimeStamp
}

/**
 * 区块的序列化方法
 */
func (block *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(block)
	if err != nil {
		panic(err.Error())
	}
	return result.Bytes()
}

/**
 * 把byte数据反序列化为区块Block结构
 */
func DeserializeBlock(blockBytes []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))
	err := decoder.Decode(&block)
	if err != nil {
		panic(err.Error())
	}
	return &block
}

/**
 * 1.创建新的区块
 */
func NewBlock(txs []*transaction.Transaction, height int64, preHash []byte) *Block {

	//1.创建新区块
	block := &Block{Height: height, PreHash: preHash, Txs: txs, TimeStamp: time.Now().Unix(), Hash: nil, Nonce: 0}

	//2.调用工作量证明对象并返回符合要求的nonce值
	pow := pow.NewProofOfWork(block)
	hash, nonce := pow.Run()

	fmt.Println()

	block.Hash = hash
	block.Nonce = nonce

	return block
}

/**
 * 2.创建创世区块
 */
func CreateGenesisBlock(txs []*transaction.Transaction) *Block {
	return NewBlock(txs, 1, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
}
