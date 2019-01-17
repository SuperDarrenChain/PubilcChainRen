package p2p

/**
 *
 */
type Version struct {
	Version    int    //当前程序版本
	LastHeight int64  //当前最新区块高度
	AddFrom    string //节点的地址
}
