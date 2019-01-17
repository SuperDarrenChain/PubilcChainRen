package p2p

import (
	"PubChainRen/block"
	"PubChainRen/utils"
	"bytes"
	"encoding/gob"
	"log"
	"fmt"

	"github.com/boltdb/bolt"
)

type Blocks struct {
	Blocks []*block.Block
}

func (blocks *Blocks) Searlize() []byte {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(blocks)
	if err != nil {
		log.Panic(err)
	}
	return buf.Bytes()
}

/**
 * 其他节点处理主节点返回的数据
 */
func HandleSendblocks(request []byte) {

	getBlocksBytes := request[utils.COMMANDLENGTH:]

	//进行反序列化
	var blocks Blocks
	reader := bytes.NewReader(getBlocksBytes)
	decoder := gob.NewDecoder(reader)
	err := decoder.Decode(&blocks)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("其他节点接收到全节点返回的区块数据:", blocks.Blocks)

	if len(blocks.Blocks) != 0 { //插入本地数据库
		fmt.Println(" 将返回的数据插入本地数据库 ")
		err := chain.DB.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(block.BlockTableName))
			if bucket == nil {
				bucket, err = tx.CreateBucket([]byte(block.BlockTableName))
				if err != nil {
					panic(err.Error())
				}
			}

			var lastHash []byte
			var lastHeight int64
			lastHeight = 0
			for _, blk := range blocks.Blocks {
				bucket.Put(blk.Hash, blk.Serialize())
				if blk.Height > lastHeight {
					lastHeight = blk.Height
					lastHash = blk.Hash
				}
			}
			//最后更新最新的数据
			err := bucket.Put([]byte(block.LastBlockByte), lastHash)
			if err != nil {
				panic(err.Error())
			}
			return nil
		})
		if err != nil {
			panic(err.Error())
		}
	}
}
