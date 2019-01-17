package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"encoding/gob"
	"log"
	"fmt"
)

//当前程序的版本
const NODE_VERSION = 1

//命令的固定长度
const COMMANDLENGTH = 12

/**
 * 将int64类型转换为字节数组
 */
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		panic(err)
	}
	return buff.Bytes()
}

/**
 * Json转字符串数组Array
 */
func JSONToStringArray(jsonstring string) []string {
	var sArr []string
	err := json.Unmarshal([]byte(jsonstring), &sArr)
	if err != nil {
		panic(err.Error())
	}
	return sArr
}

/**
 * Json转整形数组Array
 */
func JSONToIntArray(jsonStr string) []int {
	var iArr []int
	err := json.Unmarshal([]byte(jsonStr), &iArr)
	if err != nil {
		panic(err.Error())
	}
	return iArr
}

//将对象进行序列化
func GobEncode(data interface{}) []byte {
	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(data)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

/**
 * 命令转换成byte数组
 */
func CommandToBytes(command string) []byte {
	var bytes [COMMANDLENGTH]byte //定义一个长度为12的数组
	//遍历每个命令字符
	for i, c := range command {
		bytes[i] = byte(c) //将命令字符转成byte赋值给数组，[v,e,r,s,i,o,n,0,0,0,0,0]
	}
	return bytes[:]
}

/**
 * byte数组转换成命令
 */
func BytesToCommand(bytes []byte) string {
	var command []byte //用来存储拿到的命令
	for _, b := range bytes { //遍历命令的12个字节，把0删除，只保留前面的命令
		if b != 0x00 {
			command = append(command, b)
		}
	}
	//返回拿到的命令
	return fmt.Sprintf("%s", command)
}
