package wallet

import (
	"PubChainRen/utils"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"

	"bytes"
)

/**
 * 钱包结构体
 */
type Wallet struct {
	/**
	 * 私钥
	 */
	PrivateKey ecdsa.PrivateKey

	/**
	 * 公钥
	 */
	PublickKey []byte
}

const Version = byte(0x00)
const AddressChecksumLen = 4

/**
 * 判断传入的地址是否是有效的正确的地址
 */
func IsAddressValid(address string) bool {

	addressbytes := []byte(address)

	//1.base58解码
	base58Bytes := utils.Base58Decode(addressbytes)

	//2.获取地址中携带的4位校验码
	checkSumBytes := base58Bytes[len(base58Bytes)-AddressChecksumLen:]

	//3.version + 公钥哈希
	versionBytes := base58Bytes[:len(base58Bytes)-AddressChecksumLen]
	//4.重新生成校验码

	reCheckSumbytes := CheckSum(versionBytes)
	return bytes.Compare(checkSumBytes, reCheckSumbytes) == 0
}

func (wallet *Wallet) GetAddress() string {
	//1.sha256计算，ripemd160计算
	ripemd160Hash := Ripemd160Hash(wallet.PublickKey)
	version_ripemd160Hash := append([]byte{Version}, ripemd160Hash...)

	//2.checkNum
	checkSumBytes := CheckSum(version_ripemd160Hash)

	//3.拼接
	bytes := append(version_ripemd160Hash, checkSumBytes...)

	return fmt.Sprintf("%s", utils.Base58Encode(bytes))
}


/**
 * 封装获取到验证码
 */
func CheckSum(ripemd160hash []byte) []byte {
	hasher := sha256.New()
	hasher.Write(ripemd160hash)
	hasher1 := hasher.Sum(nil)

	hasher2 := sha256.New()
	hasher2.Write(hasher1)

	hasher3 := hasher2.Sum(nil)
	return hasher3[:AddressChecksumLen]
}

func Ripemd160Hash(pubKey []byte) []byte {
	//1.256hash
	hash256 := sha256.New()
	hash256.Write(pubKey)
	hash := hash256.Sum(nil)

	//2.160
	hash160 := ripemd160.New()
	hash160.Write(hash)

	return hash160.Sum(nil)
}

func NewWallet() *Wallet {
	privateKey, pubKey := newKeyPair()
	return &Wallet{privateKey, pubKey}
}

/**
 * 生成公私钥对
 */
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		panic(err.Error())
	}
	pub := append(private.X.Bytes(), private.Y.Bytes()...)
	return *private, pub
}
