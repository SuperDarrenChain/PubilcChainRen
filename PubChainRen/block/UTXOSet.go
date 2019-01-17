package block

import (
	"PubChainRen/transaction"
	"PubChainRen/wallet"
	"github.com/boltdb/bolt"
	"encoding/hex"

	"log"
	"bytes"

)

/**
 * 操作TxOutputs并存储到数据库表的结构体对象
 */
type UTXOSet struct {
	BlockChain *BlockChain
}

const UTXOSetName = "utxoset"

/**
 * 更新send
 */
func (utxoSet *UTXOSet) Update() {

	//以上内容是找出最新区块中的所有的花费了的input
	//下面开始获取最新区块中所有的output
	outsMap := make(map[string]*transaction.TxOutputs)

	//1.获取最后一个区块，遍历该区块中的所有tx
	newBlock := utxoSet.BlockChain.NewBlockIterator().Next()

	//2.获取所有的input
	inputs := []*transaction.TxInput{} //用来装所有的获取的所有input
	for _, tx := range newBlock.Txs { //遍历所有的交易
		if !tx.IsCoinbaseTransaction() { //判断不是coinbase
			for _, in := range tx.TxInputs { //遍历交易中的所有input
				inputs = append(inputs, in) //把input装到inputs中
			}
		}
	}

	//3.获取最新区块中所有的output,如果和inputs中的input对上了，就说明花了
	for _, tx := range newBlock.Txs {
		//用来装未花费的utxo
		utxos := []*transaction.UTXO{}
		//找出所有交易中的未花费
		for index, output := range tx.TxOutputs {
			isSpent := false //设已花费为false
			for _, input := range inputs { //遍历所有inputs
				//判断input中的TxID==tx.TxID && 当前input的vout是否引用的tx中的vouts中的某个output
				if bytes.Compare(tx.TxHash, input.TxHash) == 0 && index == input.Vout {
					//判断output中的pubKeyHash和input中的PubKeyHash一样的话就对上了，就表示花掉了
					version_pubKey := append(input.PubKey, wallet.Version)
					if bytes.Compare(output.PubKeyHash, wallet.Ripemd160Hash(version_pubKey)) == 0 {
						isSpent = true
					}
				}
			}

			//遍历完inputs后如果isSpent没有被标记为true。就说明Vouts中的output都是没被花费的。就全加到utxos中
			if isSpent == false {
				utxo := &transaction.UTXO{tx.TxHash, index, output}
				utxos = append(utxos, utxo)
			}
		}
		//如果utxos中有数据，就加到map中
		if len(utxos) > 0 {
			txIDStr := hex.EncodeToString([]byte(tx.TxHash))
			outsMap[txIDStr] = &transaction.TxOutputs{utxos}
		}
	}
	//以上为拿到所有的未花费的utxo到map中
	//删除花费了的数据
	err := utxoSet.BlockChain.DB.Update(func(tx *bolt.Tx) error {
		if b := tx.Bucket([]byte(UTXOSetName)); b != nil {
			//先遍历inputs，和utxoset表对比
			for _, input := range inputs {
				//从表中拿到和input对应的要查询的数据txOutputs
				txOutputsBytes := b.Get([]byte(input.TxHash))
				//判断如果拿到数据长度为0，说明没找到对应的数据，直接跳过这次循环，继续查找下一个input
				if len(txOutputsBytes) == 0 {
					continue
				}

				//如果拿到了数据，就将txOutputs反序列化
				txOutputs := transaction.DeserializeTxOutputs(txOutputsBytes)
				//是否需要被删除标记
				isNeedDelete := false

				//存储该txOutput中未花费的utxo
				utxos := []*transaction.UTXO{}

				//遍历反序列化后的所有utxo
				for _, utxo := range txOutputs.UTXOs {
					//检查pubkeyhash一样，并且下标对上了就说明花掉了。需要删除
					if bytes.Compare(utxo.Output.PubKeyHash, wallet.Ripemd160Hash(input.PubKey)) == 0 && input.Vout == utxo.Vout {
						isNeedDelete = true
					} else {
						//拿到所有的不需要删除的utxo存到utxos数组中，等待更新到库中
						utxos = append(utxos, utxo)
					}
				}
				//如果有需要删除的数据
				if isNeedDelete == true {
					b.Delete(input.TxHash) //删除input对应的那个utxo
					if len(utxos) > 0 { //如果utxos中有未花费的，需要存上
						//创建TxOutputs对象，将utxos扔里面
						txOutputs := &transaction.TxOutputs{utxos}
						//将最新的txOutputs存进去
						b.Put(input.TxHash, txOutputs.Serialize())
					}
				}
			}
			//然后将最新区块中的未花费的也存到库中
			for txIDStr, txOutputs := range outsMap {
				txID, _ := hex.DecodeString(txIDStr) //string to []byte
				b.Put(txID, txOutputs.Serialize())   //序列化txOutputs存进去
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

}

/**
 * 查询可以花费的utxos
 */
func (utxoSet *UTXOSet) FindSpentableUTXOs(from string, amount int, txs []*transaction.Transaction) (int, map[string][]int) {
	var total int //存储from的余额

	spentableUTXOMap := make(map[string][]int)

	//1.遍历未打包的交易中的可以花费的utxos
	unPackageSpentableUTXOs := utxoSet.FindUnpackeSpentableUTXO(from, txs)

	for _, utxo := range unPackageSpentableUTXOs {
		total += utxo.Output.Value
		txIDStr := hex.EncodeToString([]byte(utxo.TxHash))
		//将需要转帐的钱存到map中
		spentableUTXOMap[txIDStr] = append(spentableUTXOMap[txIDStr], utxo.Vout)
		if total > amount { //如果转帐的钱够用了，就直接返回拿到的钱
			return total, spentableUTXOMap
		}
	}

	//2.未打包的交易中遍历完毕，还不够转账的花费，则遍历utxoset表
	err := utxoSet.BlockChain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(UTXOSetName)) //查表
		if b != nil {
			//查询
			c := b.Cursor()
			//遍历数据库中所有的txoutputs
		dbLoop:
			for k, v := c.First(); k != nil; k, v = c.Next() {

				txOutputs := transaction.DeserializeTxOutputs(v)

				for _, utxo := range txOutputs.UTXOs {

					if utxo.Output.UnLockScriptPubKeyWithAddress(from) {

						total += utxo.Output.Value
						txHashStr := hex.EncodeToString(utxo.TxHash)
						//将拿到的钱都放到map中
						spentableUTXOMap[txHashStr] = append(spentableUTXOMap[txHashStr], utxo.Vout)

						if total >= amount { //如果钱够用，跳出最外层循环
							break dbLoop
						}
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	//3.返回拿到的余额和可以花的钱
	return total, spentableUTXOMap
}

/**
 * 查询未打包的tx中，可以使用的utxo
 */
func (utxoSet *UTXOSet) FindUnpackeSpentableUTXO(from string, txs []*transaction.Transaction) []*transaction.UTXO {
	//存储可以使用的未花费的utxo
	var unUTXOs []*transaction.UTXO

	//存储已经花费的input
	spentedMap := make(map[string][]int)

	//倒序遍历每个未打包的交易，去取得每个交易中的未花费的utxo
	for i := len(txs) - 1; i >= 0; i-- {
		unUTXOs = caculate(txs[i], from, spentedMap, unUTXOs)
	}
	//返回utxos
	return unUTXOs
}

/**
 * 重置utxoSet数据表
 */
func (utxoset *UTXOSet) ResetUTXOSet() {

	err := utxoset.BlockChain.DB.Update(func(tx *bolt.Tx) error {

		//1.判断bucket 是否存在
		bucket := tx.Bucket([]byte(UTXOSetName))

		if bucket != nil {
			err := tx.DeleteBucket([]byte(UTXOSetName))
			if err != nil {
				panic(err.Error())
			}
		}

		//2.创建
		bucket, err := tx.CreateBucket([]byte(UTXOSetName))
		if err != nil {
			panic(err.Error())
		}

		if bucket != nil {
			//3.将表数据存储到utxoset
			unUTXOMap := utxoset.BlockChain.FindUnSpentUTXOMap()
			//遍历拿到的所有未花费的txOutputs
			for txIDStr, outs := range unUTXOMap {
				txID, _ := hex.DecodeString(txIDStr) //字符串转成[]byte
				//将txOutputs序列化后存储到表中
				bucket.Put(txID, outs.Serialize())
			}
		}
		return nil
	})
	if err != nil {
		panic(err.Error())
	}
}

/**
 *返回某个地址的余额
 */
func (utxoset *UTXOSet) GetBalance(address string) int {
	//去utxoset中查询所有未花费的utxo
	utxos := utxoset.FindUnSpentUTXOsByAddress(address)
	var total int //用来记录余额
	for _, utxo := range utxos {
		//累加所有的未花费的金额
		total += utxo.Output.Value
	}
	//返回找到的金额
	return total
}

/**
 * 查询某个地址的所有的未花费的utxo
 */
func (utxoSet *UTXOSet) FindUnSpentUTXOsByAddress(address string) []*transaction.UTXO {
	var utxos []*transaction.UTXO
	err := utxoSet.BlockChain.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(UTXOSetName)) //打开utxoset表
		if b != nil {
			//获取表中的所有的utxo
			c := b.Cursor()
			//遍历数据库,拿到对应address的所有的txInputs
			for k, v := c.First(); k != nil; k, v = c.Next() {
				//将每一个txInoutputs反序列化
				txOutputs := transaction.DeserializeTxOutputs(v)
				//遍历反序列化后的所有的utxo
				for _, utxo := range txOutputs.UTXOs {
					//判断是否本人查询
					if utxo.Output.UnLockScriptPubKeyWithAddress(address) {
						//如果是本人查询就把查到的utxos返回
						utxos = append(utxos, utxo)
						//fmt.Println(utxo)
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		panic(err.Error())
	}
	return utxos
}
