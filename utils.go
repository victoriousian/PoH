package main

import (
	"bytes"
	"encoding/binary"
	"log"
)

func IntToHex(i int64) []byte {
	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.BigEndian, i)
	if err != nil {
		log.Panic(err)
	}

	return buf.Bytes()
}