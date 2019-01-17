package transaction

import (
	"PubChainRen/utils"
	"PubChainRen/wallet"
	"bytes"
)

/**
 *交易的输出项
 */
type TxOutput struct {
	//未花费的金额
	Value int
	//地址（属于谁），锁定脚本
	//ScriptPubKey string
	//公钥hash
	PubKeyHash []byte
}

func (txOutput *TxOutput) UnLockScriptPubKeyWithAddress(address string) bool {
	addressBytes := utils.Base58Decode([]byte(address)) //解码钱包地址
	//取出版本号+公钥哈希
	pubKeyHash := addressBytes[:len(addressBytes)-wallet.AddressChecksumLen]
	//拿着取出来的公钥哈希和txOutput中的PubKeyHash对比。如果一样，返回真
	return bytes.Compare(pubKeyHash, txOutput.PubKeyHash) == 0
}

//根据地址创建一个output对象
func NewTxOutput(value int, address string) *TxOutput {
	txOutput := &TxOutput{value, nil}
	txOutput.Lock(address)
	return txOutput
}

//利用地址的公钥hash,将币锁定在某个地址上面
func (tx *TxOutput) Lock(address string) {
	addressBytes := utils.Base58Decode([]byte(address))
	tx.PubKeyHash = addressBytes[:len(addressBytes)-wallet.AddressChecksumLen]
}
