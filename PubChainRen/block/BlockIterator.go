package block

import "github.com/boltdb/bolt"

/**
 * 区块链迭代器
 */
type BlockIterator struct {
	CurrentHash []byte //当前区块hash
	Db          *bolt.DB
}

/**
 * 创建返回一个区块链迭代器
 */
func (chain *BlockChain) NewBlockIterator() *BlockIterator {
	return &BlockIterator{chain.Tip, chain.DB}
}

/**
 * 迭代器获得下一个区块
 */
func (iterator *BlockIterator) Next() *Block {
	var block *Block
	err := iterator.Db.View(func(tx *bolt.Tx) error {
		//1。获取到表
		bucket := tx.Bucket([]byte(BlockTableName))
		if bucket != nil {
			blockBytes := bucket.Get(iterator.CurrentHash)
			if blockBytes != nil {
				block = DeserializeBlock(blockBytes)
				//修改迭代器里面的CurrentHash
				iterator.CurrentHash = block.PreHash
			} else { //否则直接返回nil
				return nil
			}
		}
		return nil
	})
	if err != nil {
		panic(err.Error())
	}
	return block
}
