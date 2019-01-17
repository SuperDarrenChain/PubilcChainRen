package transaction

/**
 * 未花费的交易类型
 */
type UTXO struct {
	//交易的hash
	TxHash []byte
	//未交易的输出在交易中的下标位置
	Vout int
	//具体的未花费的输出
	Output *TxOutput
}

