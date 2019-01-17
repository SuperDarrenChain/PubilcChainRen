package transaction

import (
	"PubChainRen/wallet"
	"bytes"
	"encoding/gob"
	"crypto/sha256"
	"encoding/hex"

	"crypto/ecdsa"
	"crypto/rand"
	"log"
	"math/big"
	"crypto/elliptic"
)

/**
 * 交易结构体
 */
type Transaction struct {
	/**
	 * UTXO模型
	 */
	TxHash []byte //交易的hash值
	//交易输入
	TxInputs []*TxInput
	//交易输出
	TxOutputs []*TxOutput
}

/**
 * 对交易进行验证
 */
func (tx *Transaction) Verifity(prevTxs map[string]*Transaction) bool {
	if tx.IsCoinbaseTransaction() {
		return true
	}
	//遍历当前传过来的要验证的交易
	for _, input := range prevTxs {
		if prevTxs[hex.EncodeToString(input.TxHash)] == nil {
			log.Panic("当前的input同有找到对应的Transaction,无法验证。。")
		}
	}
	//验证
	txCopy := tx.TrimmedCopy() //拷贝交易副本
	curev := elliptic.P256()   //创建椭圆曲线加密算法

	//遍历当前交易中的每一笔input数据
	for index, input := range tx.TxInputs {

		prevTx := prevTxs[hex.EncodeToString(input.TxHash)]
		txCopy.TxInputs[index].Signature = nil
		txCopy.TxInputs[index].PubKey = prevTx.TxOutputs[input.Vout].PubKeyHash
		txCopy.TxHash = txCopy.NewTxHash()
		txCopy.TxInputs[index].PubKey = nil

		//获取公钥,将公钥切成x,y,各32bit
		x := big.Int{}
		y := big.Int{}
		keyLen := len(input.PubKey) //获取input.PublicKey长度值，方便切
		x.SetBytes(input.PubKey[:keyLen/2])
		y.SetBytes(input.PubKey[keyLen/2:])

		//再次生成公钥哈希
		rawPublicKey := ecdsa.PublicKey{curev, &x, &y}

		//获取签名，r,s
		r := big.Int{}
		s := big.Int{}
		signLen := len(input.Signature) //获取签名的长度，方便切
		r.SetBytes(input.Signature[:signLen/2])
		s.SetBytes(input.Signature[signLen/2:])
		//验证签名
		if ecdsa.Verify(&rawPublicKey, txCopy.TxHash, &r, &s) == false {
			return false
		}
	}
	return true
}

/**
 * 将交易进行序列化
 */
func (tx *Transaction) Serialize() []byte {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	return buf.Bytes()
}

/**
 * 将当前交易生成hash
 */
func (tx *Transaction) NewTxHash() []byte {
	txCopy := tx             //将当前交易生成一个副本
	txCopy.TxHash = []byte{} //将副本中的txHash置空
	//生成hash
	hash := sha256.Sum256(txCopy.Serialize())
	//返回hash
	return hash[:]

}

//签名
/*
签名：为了对一笔交易进行签名
私钥：
要获取交易的Input,引用的output,所在的之前的交易
*/
func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey, prevTxsmap map[string]*Transaction) {
	//1.判断当前的tx是否是coinbase交易
	if tx.IsCoinbaseTransaction() {
		return
	}

	//2.当前的transaction中存储了txhash对应的output的map数组
	// 获取当前的txs中的所有的input对应的output所在的tx存不存在，如果不存在，无法进行签名
	for _, input := range tx.TxInputs { //遍历当前交易所有的Vins
		if prevTxsmap[hex.EncodeToString(input.TxHash)] == nil {
			log.Panic("当前的input，没有找到对应的output所在的Transaction,无法签名。。")
		}
	}

	//重新构建一份要签名的副本数据
	txCopy := tx.TrimmedCopy()

	//遍历拿到的副本中的数据进行遍历
	for index, input := range txCopy.TxInputs {
		//从map中拿到当前input对应的tx
		pervTx := prevTxsmap[hex.EncodeToString(input.TxHash)]

		input.Signature = nil

		input.PubKey = pervTx.TxOutputs[input.Vout].PubKeyHash

		//开始签名
		/*
		1.第一个参数是随机内存数
		2.第二个参数是参数传过来的私钥
		3.第三个参数是将设置好的txCopy做sha256获取到的hash数据
		*/
		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.NewTxHash())
		if err != nil {
			log.Panic(err)
		}
		//拼接r+s,就拿到了签名
		sign := append(r.Bytes(), s.Bytes()...)

		input.PubKey = nil

		tx.TxInputs[index].Signature = sign
	}
}

//获取要签名的tx的副本
//要签名的tx中，并不是所有数据都要作为签名数据只是一部分
/*
需要的数据如下
TxID
Inputs
	txhash,index
Outputs
	value,pubkeyhash
注意，除了Inputs中的sign,publickey不要以外，其它都要
*/
//属于tx的方法，处理完副本后返回处理好的tx副本数据
func (tx *Transaction) TrimmedCopy() *Transaction {

	var inputs []*TxInput   //用于存储所有的TxInput
	var outputs []*TxOutput //用于存储所有的TxOutput

	for _, in := range tx.TxInputs { //遍历所有的input
		//将所有的input的signature和PublicKey置空，然后追加到inputs中，这样就拿到了所有的处理好的TxInput
		inputs = append(inputs, &TxInput{in.TxHash, in.Vout, nil, nil})
	}
	for _, out := range tx.TxOutputs { //遍历所有的output
		outputs = append(outputs, &TxOutput{out.Value, out.PubKeyHash})
	}

	//创建新的transaction
	txCopy := &Transaction{tx.TxHash, inputs, outputs}

	return txCopy
}

/**
 * 创建新的交易
 */
func NewTransaction(from, to string, amount int, totalValue int, spentableUtxo map[string][]int, node_id string) *Transaction {

	var txInputs []*TxInput
	var txOutputs []*TxOutput

	//1.根据from返回所有的未花费的交易输出所对应的transaction
	//unSpentTx := UTXOWithAddress(from)
	//fmt.Println(utxos)

	//2.根据已经查询到的包含address未花费的交易输出的所有交易,统计出来所有可以花费的总数,及余额所在的交易和下标
	//txHash : 7b0e3b3f258727db14703895ed5fff542dd82526fb2e4ab9696e86edbe2395ea

	//获取钱包的集合
	wallets := wallet.NewWallets(node_id)
	wallet := wallets.Wallets[from]

	//构建输入
	for txhash, vouts := range spentableUtxo {
		bytes, _ := hex.DecodeString(txhash)
		for _, vout := range vouts {
			input := &TxInput{bytes, vout, nil, wallet.PublickKey}
			txInputs = append(txInputs, input)
		}
	}

	//构建输出
	output1 := NewTxOutput(amount, to)
	txOutputs = append(txOutputs, output1)
	if totalValue-amount > 0 { //只有需要找零时才创建找零output
		output2 := NewTxOutput(totalValue-amount, from)
		txOutputs = append(txOutputs, output2)
	}

	transaction := &Transaction{nil, txInputs, txOutputs}
	transaction.SetHashTransaction()

	return transaction
}

/**
 * 判断某个交易是否是coinbase交易
 */
func (tx *Transaction) IsCoinbaseTransaction() bool {
	var hashInt big.Int
	hashInt.SetBytes(tx.TxInputs[0].TxHash)
	return big.NewInt(0).Cmp(&hashInt) == 0 && tx.TxInputs[0].Vout == -1
}

/**
 * Coinbase交易：区块的第一笔交易，只有输出，没有输入
 */
func NewCoinbaseTransaction(address string) *Transaction {
	//输入

	txInput := &TxInput{[]byte{0}, -1, nil, nil}
	//输出
	txOutput := NewTxOutput(10, address)
	//txOutput := &TxOutput{10, address}

	txCoinbase := &Transaction{[]byte{}, []*TxInput{txInput}, []*TxOutput{txOutput}}
	//设置交易hash
	txCoinbase.SetHashTransaction()

	return txCoinbase
}

/**
 * 对Transaction进行hash计算，得到交易hash
 */
func (tx *Transaction) SetHashTransaction() {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(tx)
	if err != nil {
		panic(err.Error())
	}
	hash := sha256.Sum256(result.Bytes())
	tx.TxHash = hash[:]
}
