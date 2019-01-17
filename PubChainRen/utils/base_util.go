package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"math/big"
	"strconv"
	"strings"
)

//最常用的函数

/**
 * 将字节数组转成16进制字符串： []byte -> string
 */
func BytesToHexString(arr []byte) string {
	return hex.EncodeToString(arr)
}

/**
 * 将16进制字符串转成字节数组： hex string ->  []byte
 */
func HexStringToBytes(s string) ([]byte, error) {
	arr, err := hex.DecodeString(s)
	return arr, err
}

/**
 * 16进制字符串大端和小端颠倒
 */
func ReverseHexString(hexStr string) string {
	arr, _ := hex.DecodeString(hexStr)
	ReverseBytes(arr)
	return hex.EncodeToString(arr)
}

/**
 * 字节数组大端和小端颠倒
 */
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

//截取字符串 start 起点下标 end 终点下标(不包括)
func Substring(str string, start int, end int) string {
	rs := []rune(str)
	length := len(rs)

	if start < 0 || start > length {
		panic("start is wrong")
	}

	if end < 0 || end > length {
		panic("end is wrong")
	}

	return string(rs[start:end])
}

func Substr2(str string, len int) string {
	rs := []rune(str)
	result := string(rs[:len])
	return result
}

func Substr(str string, start int, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0

	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}

	return string(rs[start:end])
}

//对称加密需要的填充函数
func PKCS5Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

func PKCS5UnPadding(data []byte) []byte {
	// 去掉最后一个字节 unpadding 次
	unpadding := int(data[len(data)-1])
	return data[:(len(data) - unpadding)]
}

func ZeroPadding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(data, padtext...)
}

func ZeroUnPadding(data []byte) []byte {
	return bytes.TrimRightFunc(data, func(r rune) bool {
		return r == rune(0)
	})
}

/*
不常用的函数
*/
//10进制数值转16进制
func DecimalToHex(n int64) string {
	if n < 0 {
		log.Println("Decimal to hexadecimal error: the argument must be greater than zero.")
		return ""
	}
	if n == 0 {
		return "0"
	}
	hex := map[int64]int64{10: 65, 11: 66, 12: 67, 13: 68, 14: 69, 15: 70}
	s := ""
	for q := n; q > 0; q = q / 16 {
		m := q % 16
		if m > 9 && m < 16 {
			m = hex[m]
			s = fmt.Sprintf("%v%v", string(m), s)
			continue
		}
		s = fmt.Sprintf("%v%v", m, s)
	}
	return s
}

// Hexadecimal to decimal
func HexStringToDec(h string) (n int64) {
	s := strings.Split(strings.ToUpper(h), "")
	l := len(s)
	i := 0
	d := float64(0)
	hex := map[string]string{"A": "10", "B": "11", "C": "12", "D": "13", "E": "14", "F": "15"}
	for i = 0; i < l; i++ {
		c := s[i]
		if v, ok := hex[c]; ok {
			c = v
		}
		f, err := strconv.ParseFloat(c, 10)
		if err != nil {
			log.Println("Hexadecimal to decimal error:", err.Error())
			return -1
		}
		d += f * math.Pow(16, float64(l-i-1))
	}
	return int64(d)
}

//16进制长字符串转10进制字符串
func HexStringToDecimal(hexStr string) string {
	//bInt := big.Int{}
	bInt := new(big.Int)
	bytes, _ := hex.DecodeString(hexStr)
	bInt.SetBytes(bytes)
	return bInt.Text(10)
}

//16进制长字符串转长整数
func HexStringToBigint(hexStr string) *big.Int {
	bInt := big.Int{}
	bytes, _ := hex.DecodeString(hexStr)
	bInt.SetBytes(bytes)
	return &bInt
}

////16进制长字符串转长整数
//func DecimalStringToBigint(decStr string) *big.Int {
//	//bInt := big.Int{}
//	return big.NewInt(int(decStr))
//}

/**
 * int64转字节数组:  int64 -> []byte
 */
func IntToBytes(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

func BytesToString(b []byte) (s string) {
	s = ""
	for i := 0; i < len(b); i++ {
		s += fmt.Sprintf("%02X", b[i])
	}
	return s
}
