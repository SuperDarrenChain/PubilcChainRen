package pow

import (
	"PubChainRen/utils"
	"math/big"
	"crypto/sha256"
	"bytes"

	"fmt"
)

// 区块Hash：0000 0000 0019 d668 9c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f 共64位
//   二进制：0000 0000 0000 0000  0000 0000 0000 0000  0000 0000 0000 0000  0000 0000 0000 0000 ....

//

const TargetBit = 16

/**
 *工作量证明结构体
 */
type ProofOfWork struct {
	Block  Block    //当前要验证的区块
	Target *big.Int //大数据存储
}

/**
 * block的接口
 */
type Block interface {
	GetHeight() int64
	GetPreHash() []byte
	GetTxsHash() []byte
	GetTimeStamp() int64
}

/**
 * 1.创建工作量证明实体
 */
func NewProofOfWork(block Block) *ProofOfWork {
	//1.big.Int对象 1

	//1.创建一个初始值为1的target
	target := big.NewInt(1)
	//2.左移256 - targetBit
	target = target.Lsh(target, 256-TargetBit)

	return &ProofOfWork{block, target}
}

/**
 * 2.执行run方法，对某一个区块进行工作量证明验证和计算，并返回符合要求的nonce值和hash
 */
func (pow *ProofOfWork) Run() ([]byte, int64) {
	//1.将block的属性拼接成字节数组
	//2.生成hash
	//3.判断hash是否有效，满足要求，停止循环

	var nonce int64
	nonce = 0

	var hash [32]byte
	var hashInt big.Int

	for {
		dataBytes := pow.PrepareData(nonce)

		hash = sha256.Sum256(dataBytes)
		fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])
		// Cmp compares x and y and returns:
		//
		//   -1 if x <  y
		//    0 if x == y
		//   +1 if x >  y
		//
		if pow.Target.Cmp(&hashInt) == 1 {
			break
		}
		nonce++
	}
	return hash[:], nonce
}

/**
 * 准备hash计算数据
 */
func (pow *ProofOfWork) PrepareData(nonce int64) []byte {
	blockBytes := bytes.Join([][]byte{
		utils.IntToHex(pow.Block.GetHeight()),
		pow.Block.GetPreHash(),
		pow.Block.GetTxsHash(),
		utils.IntToHex(pow.Block.GetTimeStamp()),
		utils.IntToHex(nonce)}, []byte{})
	return blockBytes
}
