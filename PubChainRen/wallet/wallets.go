package wallet

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"crypto/elliptic"
	"os"
	"fmt"
)

/**
 * 钱包集合结构体
 */
type Wallets struct {
	Wallets map[string]*Wallet
}

const walletFile = "wallets_%s.dat"

/**
 * 钱包管理工具创建钱包
 */
func (wallets *Wallets) CreateWallet() {
	wallet := NewWallet()
	address := wallet.GetAddress()
	fmt.Println(address)
	wallets.Wallets[address] = wallet
}

/**
 * 创建钱包集合对象
 */
func NewWallets(node_id string) *Wallets {
	//从本地数据文件中获取钱包信息
	wallets, err := LoadFromFile(node_id)
	if err != nil {
		panic(err.Error())
	}

	return wallets
}

/**
 * 从本地文件读取数据
 */
func LoadFromFile(node_id string) (*Wallets, error) {
	walletsFile := fmt.Sprintf(walletFile, node_id)
	if _, err := os.Stat(walletsFile); os.IsNotExist(err) {
		wallets1 := &Wallets{}
		wallets1.Wallets = make(map[string]*Wallet)
		return wallets1, nil
	}

	fileContent, err := ioutil.ReadFile(walletsFile)
	if err != nil {
		panic(err.Error())
	}
	var wallets Wallets
	//反序列化之前注册涉及到的数据类型
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		panic(err.Error())
	}
	return &wallets, nil
}

/**
 * 保存wallets数据
 */
func (wallets *Wallets) SaveToFile(node_id string) {
	var content bytes.Buffer

	//注册elliptic,目的：可以序列化elliptic中的接口及数据类型
	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(wallets)
	if err != nil {
		panic(err.Error())
	}

	walletsFile := fmt.Sprintf(walletFile, node_id)
	//将序列化后的数据写入到文件中，并覆盖原来的内容
	err = ioutil.WriteFile(walletsFile, content.Bytes(), 0644)
	if err != nil {
		panic(err.Error())
	}
}
