package merkletree

import (
	"crypto/sha256"
	"math"
)

//创建结构体对象，表示节点和树
//这是节点
type MerkleNode struct {
	LeftNode  *MerkleNode //左子树
	RightNode *MerkleNode //右子树
	DataHash  []byte      //数据hash
}

//树,只有根节点
type MerkleTree struct {
	RootNode *MerkleNode
}

//2.给一个左右节点，生成一个新的节点
func NewMerkleNode(leftNode, rightNode *MerkleNode, txHash []byte) *MerkleNode {

	//1.创建新子节点
	mNode := &MerkleNode{}

	//2.左孩子和右孩子都为nil，及为叶子节点
	if leftNode == nil && rightNode == nil {
		hash := sha256.Sum256(txHash)
		mNode.DataHash = hash[:]
	} else { //非叶子节点，需要找到左孩子和右孩子的hash，然后再进行hash
		prevHash := append(leftNode.DataHash, rightNode.DataHash...)
		hash := sha256.Sum256(prevHash)
		mNode.DataHash = hash[:]
	}

	mNode.LeftNode = leftNode   //左儿子
	mNode.RightNode = rightNode //右儿子
	return mNode
}

//生成MerkleTree
func NewMerkleTree(txHashData [][]byte) *MerkleTree {
	//1.创建一个数组，用于存储node节点
	var nodes []*MerkleNode

	//2.判断交易量的奇偶性，如果交易的长度为奇数个，则复制最后一个交易，变成偶数
	if len(txHashData)%2 != 0 {
		//奇数，复制最后一个
		txHashData = append(txHashData, txHashData[len(txHashData)-1])
	}

	//3.创建一排叶子节点，遍历交易hash数据
	for _, datum := range txHashData {
		//叶子节点左右节点都为nil,再传进去交易tx序列化数据
		node := NewMerkleNode(nil, nil, datum)
		//将生成好的叶子节点都追加到nodes数组中
		nodes = append(nodes, node)
	}

	//4.生成树其他的节点
	count := GetCircleCount(len(nodes))

	//生成其它节点，直到生成到最后的根节点
	for i := 0; i < count; i++ {
		var newLevel []*MerkleNode
		//两两哈希
		for j := 0; j < len(nodes); j += 2 {
			//两两哈希，生成爹，爹就有左右儿子了
			node := NewMerkleNode(nodes[j], nodes[j+1], nil)
			//将生成的新节点追加到newLevel数组中
			newLevel = append(newLevel, node)
		}

		//先判断newLevel的奇偶性
		if len(newLevel)%2 != 0 { //如果是奇数
			newLevel = append(newLevel, newLevel[len(newLevel)-1]) //最后一个生成副本
		}
		nodes = newLevel
	}

	//拿到rootNode
	mTree := &MerkleTree{nodes[0]}

	//返回根节点
	return mTree
}

//统计几层
func GetCircleCount(len int) int {
	count := 0
	for {
		//计算2的几次方>=len
		if int(math.Pow(2, float64(count))) >= len {
			return count //如果找到了层数就返回
		}
		//没找到层数，就继续++
		count++
	}
}
