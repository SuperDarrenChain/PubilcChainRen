package transaction

import (
	"bytes"
	"encoding/gob"
	"log"
)

//存储所有未花费的UTXO
type TxOutputs struct {
	UTXOs []*UTXO
}

//序列化TxOutputs
func (outs *TxOutputs) Serialize() []byte {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(outs)
	if err != nil {
		log.Panic(err)
	}
	return buf.Bytes()
}

//反序列化
func DeserializeTxOutputs(data []byte) *TxOutputs {
	txOutputs := TxOutputs{} //创建一个TxOutputs空对象，用来存反序列化的
	reader := bytes.NewReader(data)
	decoder := gob.NewDecoder(reader)
	err := decoder.Decode(&txOutputs)
	if err != nil {
		log.Panic(err)
	}
	return &txOutputs
}
