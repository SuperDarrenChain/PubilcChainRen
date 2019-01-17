package transaction

import (
	"PubChainRen/utils"
	"PubChainRen/wallet"
	"bytes"

)

/**
 * 交易的输入项
 */
type TxInput struct {
	//交易hash
	TxHash []byte
	//使用txhash交易中的第一个输出
	Vout int
	//解锁脚本
	//ScriptSig string

	Signature []byte //数字签名
	PubKey    []byte //公钥,原始公钥
}

/**
 * 判断当前的要消费某个人的钱
 */
func (txInput *TxInput) UnLockWithAddress(address string) bool {
	//传递的地址生成公钥hash
	addressBytes := utils.Base58Decode([]byte(address))
	addressPubHash := addressBytes[:len(addressBytes)-wallet.AddressChecksumLen]

	//用原始公钥生成公钥哈希
	pubKeyHash2 := wallet.Ripemd160Hash(txInput.PubKey)
	version_ripemd160Hash := append([]byte{wallet.Version}, pubKeyHash2...)

	//用生成的公钥哈希和传过来的对比一下，一样就返回true
	return bytes.Compare(addressPubHash, version_ripemd160Hash) == 0
}
