package hashing

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"fmt"
)

func NewStructChecksum(v any) (string, error) {
	buf := bytes.NewBuffer(nil)
	err := gob.NewEncoder(buf).Encode(v)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5.Sum(buf.Bytes())), nil
}
