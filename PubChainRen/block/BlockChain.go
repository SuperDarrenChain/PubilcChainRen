package block

import (
	"PubChainRen/transaction"
	"PubChainRen/utils"
	"PubChainRen/wallet"
	"github.com/boltdb/bolt"
	"fmt"
	"math/big"
	"time"
	"os"
	"encoding/hex"

	"crypto/ecdsa"
	"bytes"
)

const DbName = "chain_%s.db"
const BlockTableName = "block"
const LastBlockByte = "lastHash"

/**
 *区块链结构体
 */
type BlockChain struct {
	Tip []byte   //最新的区块的hash值
	DB  *bolt.DB //存储区块数据的数据库
}

/**
 * 获取区块数据
 */
func (chain *BlockChain) GetBlocks(from int64, count int64) []*Block {

	iterator := chain.NewBlockIterator()
	//区块的map key为区块number高度 ，value为区块结构体对象
	// 1 ：block1
	// 2 : block2
	// 3 : block3
	blocksMap := make(map[int64]*Block)
	for {

		block := iterator.Next()

		if block != nil {
			blocksMap[block.Height] = block
		} else {
			return nil
		}

		//退出的操作
		var hashInt big.Int
		hashInt.SetBytes(block.PreHash)
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}

	var results []*Block
	for i := from; i <= from+count; i++ {
		blockObj := blocksMap[i]
		if blockObj != nil && blockObj.Height >= from && blockObj.Height <= from+count {
			results = append(results, blockObj)
		}
	}
	return results
}

/*
 * 获取最新区块的高度
 */
func (chain *BlockChain) GetLastHeight() int64 {
	iterator := chain.NewBlockIterator()
	block := iterator.Next()
	if block != nil {
		return block.Height
	}
	return 0
}

/**
 * 遍历所有的区块中的未花费的交易输出，并组合到map中返回
 */
func (chain *BlockChain) FindUnSpentUTXOMap() map[string]*transaction.TxOutputs {
	iterator := chain.NewBlockIterator()

	spentedMap := make(map[string][]*transaction.TxInput)

	unSpentUTXOMap := make(map[string]*transaction.TxOutputs)

	for {

		block := iterator.Next()

		for i := len(block.Txs) - 1; i >= 0; i-- {
			tx := block.Txs[i]
			txOutputs := &transaction.TxOutputs{[]*transaction.UTXO{}}

			//1.遍历inputs，coinbase交易不需要遍历input
			if !tx.IsCoinbaseTransaction() {
				for _, input := range tx.TxInputs {
					key := hex.EncodeToString(input.TxHash)
					spentedMap[key] = append(spentedMap[key], input)
				}
			}

			//2.遍历outputs
		output:
			for index, output := range tx.TxOutputs {

				txHash := hex.EncodeToString(tx.TxHash)
				inputs := spentedMap[txHash]
				if (len(unSpentUTXOMap) > 0) {
					var isSpent bool
					for _, input := range inputs {
						intputPubKey := wallet.Ripemd160Hash(input.PubKey)
						if index == input.Vout && bytes.Compare(intputPubKey, output.PubKeyHash) == 0 {
							isSpent = true
							continue output
						}
					}
					if isSpent == false {
						utxo := &transaction.UTXO{tx.TxHash, index, output}
						txOutputs.UTXOs = append(txOutputs.UTXOs, utxo)
					}
				} else {
					utxo := &transaction.UTXO{tx.TxHash, index, output}
					txOutputs.UTXOs = append(txOutputs.UTXOs, utxo)
				}
			}

			//3.将当前tx的未花费的交易utxo集合放入到map中
			key := hex.EncodeToString(tx.TxHash)
			unSpentUTXOMap[key] = txOutputs
		}

		var hashInt big.Int
		hashInt.SetBytes(block.PreHash)
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}
	return unSpentUTXOMap
}

func (chain *BlockChain) VerifityTransaction(tx *transaction.Transaction) bool {

	prevTxs := make(map[string]*transaction.Transaction)

	for _, input := range tx.TxInputs {
		//拿到对应的input.Txid对应的数据库中对应的transaction
		prevTx, err := chain.FindTransaction(input.TxHash)
		if err != nil {
			panic(err.Error())
		}
		//用当前的input.Txid作为key,拿到的transaction作为value，存储到map中
		prevTxs[hex.EncodeToString(input.TxHash)] = &prevTx
	}
	//验证
	return tx.Verifity(prevTxs)
}

/**
 * 转账时需要找到符合转账要求的金额，以及utxo，该方法用于从所有的uxtos当中计算并拿到所需要的部分utxo和金额
 */
func (chain *BlockChain) FindUTXOAndAmountFromUTXOList(address string, amount int, txs []*transaction.Transaction) (int, map[string][]int) {

	//1.获取某个地址对应的所有的utxo
	utxos := chain.UTXOWithAddress(address, txs)

	//2.从所有的utxos当中选出转账需要的一些utxo
	var value = 0 //选中的utxo的累计的金额
	usableUTXO := make(map[string][]int)

	for _, utxo := range utxos {
		//1.先进行金额累加
		value += utxo.Output.Value //1+2+6

		//2.把utxo进行存储
		txHash := hex.EncodeToString(utxo.TxHash)
		usableUTXO[txHash] = append(usableUTXO[txHash], utxo.Vout)

		//3.进行判断 累计金额如果已经大于转账金额，就直接返回
		if (value >= amount) {
			break
		}
	}
	//9                 4
	//13               15
	//进行判断 累计金额如果已经小于转账金额，提示用户，并退出程序
	if value < amount {
		fmt.Println("抱歉，余额不足，无法完成转账")
		os.Exit(0)
	}
	return value, usableUTXO
}

/**
 * 返回address所有的TxOutput未花费的交易输出所在的交易,因为一个address包含多笔未花费的交易输出,所以返回一个数组
 */
func (chain *BlockChain) UTXOWithAddress(address string, txs []*transaction.Transaction) []*transaction.UTXO {

	//未花费的交易输出
	var utxos []*transaction.UTXO
	//已经花费的交易输出的统计表
	spentTxOutputs := make(map[string][]int)

	//1.先遍历当前还未打包到区块但已经创建的交易数组
	for i := len(txs) - 1; i >= 0; i-- {
		//查询还没有生成区块的Transaction，用的倒序查询，查出来append到unSpentUTXOs中
		utxos = caculate(txs[i], address, spentTxOutputs, utxos)
	}

	//2.遍历数据库区块
	iterator := chain.NewBlockIterator()
	for {
		//获取到前一个区块
		block := iterator.Next()
		fmt.Println(block)
		//2.遍历该block的Txs
		for i := len(block.Txs) - 1; i >= 0; i-- {
			//查询数据库中的区块的Transaction，用的还是倒序查询，查出来append到unSpentUTXOs中
			utxos = caculate(block.Txs[i], address, spentTxOutputs, utxos)
		}

		//判断是否是创世区块，如果是创世区块，跳出循环
		var hashInt big.Int
		hashInt.SetBytes(block.PreHash)
		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
		fmt.Println()
	}
	return utxos
}

//查询所有未花费的output
func caculate(tx *transaction.Transaction, address string, spentTxOutputMap map[string][]int, unSpentUTXOs []*transaction.UTXO) []*transaction.UTXO {
	//遍历每个tx：txID，Vins，Vouts
	//遍历所有的TxInput
	if !tx.IsCoinbaseTransaction() { //如果tx不是CoinBase交易就遍历TxInput,否则就不用遍历TxInput
		for _, txInput := range tx.TxInputs {
			if txInput.UnLockWithAddress(address) {
				//txInput的解锁脚本(用户名) 如果和要查询的余额的用户名相同，
				key := hex.EncodeToString(txInput.TxHash)
				//将查询到的已花费的这笔钱存到map中，
				spentTxOutputMap[key] = append(spentTxOutputMap[key], txInput.Vout)
			}
		}
	}
	//遍历所有的TxOutput交易
outputs:
	for index, txOutput := range tx.TxOutputs {
		//判断遍历的txOutput是否是和要查询余额的这个人的，如果不是就不用执行下面语句
		if txOutput.UnLockScriptPubKeyWithAddress(address) {
			fmt.Println(address, index)
			if len(spentTxOutputMap) != 0 { //如果map记录了txInput，就去过滤
				var isSpentOutput bool //记录是否已花费
				for txID, indexArray := range spentTxOutputMap {
					for _, i := range indexArray {
						fmt.Println(i, index)
						if i == index && hex.EncodeToString(tx.TxHash) == txID {
							isSpentOutput = true //标记当前的txOutput已经花费掉了
							continue outputs     //当前区块的tx.Vouts数组就不用遍历了
						}
					}
				}
				//如果此txOutput没有在map中查到被花费掉，就记录一下
				if !isSpentOutput {
					utxo := &transaction.UTXO{tx.TxHash, index, txOutput}
					unSpentUTXOs = append(unSpentUTXOs, utxo)
				}
			} else {
				utxo := &transaction.UTXO{tx.TxHash, index, txOutput}
				unSpentUTXOs = append(unSpentUTXOs, utxo)
			}
		}
	}
	return unSpentUTXOs
}

/**
 * 区块链的获取某个地址余额的功能
 */
func (chain *BlockChain) GetBalance(address string) int64 {
	utxos := chain.UTXOWithAddress(address, []*transaction.Transaction{})
	var balance int64
	balance = 0
	for _, utxo := range utxos {
		balance += int64(utxo.Output.Value)
	}
	return balance
}

/**
 * 根据txHash找到output所对应的Transaction
 */
func (chain *BlockChain) FindTransaction(txhash []byte) (transaction.Transaction, error) {
	iterator := chain.NewBlockIterator()
	for {
		block := iterator.Next()
		for _, tx := range block.Txs {
			if bytes.Compare(tx.TxHash, txhash) == 0 {
				return *tx, nil
			}
		}

		var hashInt big.Int
		hashInt.SetBytes(block.PreHash)
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}
	var tx transaction.Transaction
	return tx, nil
}

/**
 * 对交易进行签名
 */
func (chain *BlockChain) SignTransaction(tx *transaction.Transaction, private ecdsa.PrivateKey) {

	if tx.IsCoinbaseTransaction() {
		return
	}

	//之前被引用的输出的交易
	prevTxs := make(map[string]*transaction.Transaction)

	for _, vin := range tx.TxInputs {
		prevTx, err := chain.FindTransaction(vin.TxHash)
		if err != nil {
			panic(err.Error())
		}
		prevTxs[hex.EncodeToString(prevTx.TxHash)] = &prevTx
	}

	tx.Sign(private, prevTxs)
}

/**
 * 发送交易
 */
func (chain *BlockChain) SendTransactions(from, to []string, amount []int, node_id string) {

	//1.通过循环遍历数组，构建交易对象

	//根据from返回所有的未花费的交易输出所对应的transaction
	//unSpentTx := chain.UTXOWithAddress(from[0])

	length := len(from)
	txs := []*transaction.Transaction{}

	//coinbase奖励 在最前面
	tx := transaction.NewCoinbaseTransaction(from[0])
	txs = append(txs, tx)
	fmt.Println(tx)

	/**
	 * 多笔转账：./main sentTransaction -from '["davie","mingxu","aze"]' -to '["mingxu","aze","davie"]' -amount '[3,2,1]'
	 */
	wallets := wallet.NewWallets(node_id)

	utxoSet := &UTXOSet{chain}
	for i := 0; i < length; i++ {
		//多笔转账交易代码逻辑分析
		//totalAmount, utxos := chain.FindUTXOAndAmountFromUTXOList(from[i], amount[i], txs)
		//从utxoset中拿钱
		totalAmount, utxos := utxoSet.FindSpentableUTXOs(from[i], amount[i], txs)
		tx := transaction.NewTransaction(from[i], to[i], amount[i], totalAmount, utxos, node_id)

		wallet := wallets.Wallets[from[i]]
		fmt.Println(wallet.GetAddress())
		//对交易进行签名
		chain.SignTransaction(tx, wallet.PrivateKey)

		txs = append(txs, tx)
	}

	//创建新区块之前验证签名有效性
	//遍历每一笔交易签名的有效性

	for _, tx := range txs {
		if chain.VerifityTransaction(tx) == false {
			panic("签名验证失败")
			os.Exit(1)
		}
	}

	//2.创建新的区块
	db := chain.DB
	var blockHash []byte
	var lastBlock *Block
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BlockTableName))
		if bucket != nil {
			lasthash := bucket.Get([]byte(LastBlockByte))
			if lasthash != nil {
				blockHash = bucket.Get(lasthash)
			}
		}
		return nil
	})
	if err != nil {
		panic(err.Error())
	}
	//反序列化得到最新的区块结构
	if blockHash != nil {
		lastBlock = DeserializeBlock(blockHash)
	}

	//创建新区块
	newBlock := NewBlock(txs, lastBlock.Height+1, lastBlock.Hash)

	//3.保存区块，更新最新的hash数据
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BlockTableName))
		if bucket == nil {
			bucket, err = tx.CreateBucket([]byte(BlockTableName))
			if err != nil {
				panic(err.Error())
			}
		}

		if bucket != nil {
			//保存最新的区块数据到数据库
			if err = bucket.Put(newBlock.Hash, newBlock.Serialize()); err != nil {
				panic(err.Error())
			}
			//更新最新的区块标志hash
			bucket.Put([]byte(LastBlockByte), newBlock.Hash)
			//更新chain的Tip
			chain.Tip = newBlock.Hash
		}
		return nil
	})
	if err != nil {
		panic(err.Error())
	}
}

/**
 * 获取BlockChain对象
 */
func GetBlockChain(node_id string) *BlockChain {
	dbName := fmt.Sprintf(DbName, node_id)

	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		fmt.Println(err.Error())
		panic(err.Error())
	}

	var blockHash []byte
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BlockTableName))
		if bucket != nil {
			lasthash := bucket.Get([]byte(LastBlockByte))
			if lasthash != nil {
				blockHash = lasthash
			}
		}
		return nil
	})
	if err != nil {
		panic(err.Error())
	}
	return &BlockChain{blockHash, db}
}

/**
 * 遍历输出所有区块的信息
 */
func (chain *BlockChain) PrintChain() {
	var block *Block
	iterator := chain.NewBlockIterator()

	for {

		block = iterator.Next()

		fmt.Printf("Height:%d\n", block.Height)
		fmt.Printf("PreHash:%x\n", block.PreHash)
		//fmt.Printf("Txs:%v\n", block.Txs)
		fmt.Printf("Timestamp:%s\n", time.Unix(block.TimeStamp, 0).Format("2006-01-02 03:04:05 PM"))
		fmt.Printf("Hash:%x\n", block.Hash)
		fmt.Printf("Nonce:%x\n", block.Nonce)
		fmt.Println("Txs:")
		for index, tx := range block.Txs {

			fmt.Printf("第%d笔交易:\n", (index + 1))
			//打印交易的哈希 txhash
			fmt.Printf("交易Hash：%x\n", tx.TxHash)

			//打印交易的输入
			fmt.Println("交易输入：")
			for _, txInput := range tx.TxInputs {
				fmt.Printf("%x，%d，%x\n", txInput.TxHash, txInput.Vout, txInput.PubKey)
			}

			fmt.Println("交易输出：")
			//打印交易的输出
			for _, txOutput := range tx.TxOutputs {
				pubKeyHash := txOutput.PubKeyHash

				//2.checkNum
				checkSumBytes := wallet.CheckSum(pubKeyHash)

				//3.拼接
				bytes := append(pubKeyHash, checkSumBytes...)

				address := fmt.Sprintf("%s", utils.Base58Encode(bytes))
				fmt.Printf("%d，%s\n", txOutput.Value, address)
			}

			fmt.Println()
		}

		var hashInt big.Int
		hashInt.SetBytes(block.PreHash)
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
		fmt.Println()
	}
}

/**
 * 判断链是否已经存在
 */
func IsChainExit(node_id string) bool {
	dbName := fmt.Sprintf(DbName, node_id)
	if _, err := os.Stat(dbName); os.IsNotExist(err) {
		return false
	}
	return true
}

/**
 *1.创建带有创世区块的区块链
 */
func CreateBlockChainWithGenesisBlock(txs []*transaction.Transaction, node_id string) {

	var blockHash []byte //最新的区块的hash

	/**
	 * 判断数据库是否存在。数据库存在，表示创世区块已经创建了就返回
	 */
	if IsChainExit(node_id) {
		fmt.Println("区块链已经创建，请勿重复创建")
		return
	}

	dbName := fmt.Sprintf(DbName, node_id)
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		panic(err.Error())
	}

	err = db.Update(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(BlockTableName))
		if bucket == nil {
			bucket, err = tx.CreateBucket([]byte(BlockTableName))
			if err != nil {
				panic(err.Error())
			}
		}

		if bucket != nil {
			//1.创建创世区块
			gensisBlock := CreateGenesisBlock(txs)
			//将创世区块存储到表中
			err = bucket.Put(gensisBlock.Hash, gensisBlock.Serialize())

			if err != nil {
				panic(err.Error())
			}
			//修改最新的区块的hash标志
			bucket.Put([]byte(LastBlockByte), gensisBlock.Hash)
			blockHash = gensisBlock.Hash
		}
		return nil
	})
	//关闭数据库
	db.Close()

	blockChain := GetBlockChain(node_id)
	defer blockChain.DB.Close()

	//重置utxoset：
	utxoSet := &UTXOSet{blockChain}
	utxoSet.ResetUTXOSet()
}

/**
 * 2.添加新区块到区块链中
 */
func (chain *BlockChain) AddBlockToBlockChain(txs []*transaction.Transaction) {

	//2.添加新区块到区块链中
	//chain.Blocks = append(chain.Blocks, newBlock)
	err := chain.DB.Update(func(tx *bolt.Tx) error {
		//1.获取表
		bucket := tx.Bucket([]byte(BlockTableName))

		//获取数据库中的最新的区块bytes
		blockBytes := bucket.Get(chain.Tip)
		lastBlock := DeserializeBlock(blockBytes)

		//2.创建新区块
		newBlock := NewBlock(txs, lastBlock.Height+1, lastBlock.Hash)

		if bucket != nil {
			//3.添加新区块数据到数据库中
			err := bucket.Put(newBlock.Hash, newBlock.Serialize())
			if err != nil {
				panic(err.Error())
			}

			//4.更新数据库中的最新区块bytes字段
			err = bucket.Put([]byte(LastBlockByte), newBlock.Hash)

			//5.更新blockchain的tip字段
			chain.Tip = newBlock.Hash

			if err != nil {
				panic(err.Error())
			}
		}
		return nil
	})

	if err != nil {
		panic(err.Error())
	}
}
