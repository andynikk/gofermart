package cryptography

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"

	"github.com/andynikk/gofermart/internal/constants"
)

func HeshSHA256(data string, strKey string) (hash string) {

	if strKey == "" {
		strKey = string(constants.HashKey[:])
	}

	key := []byte(strKey)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	hash = fmt.Sprintf("%x", h.Sum(nil))
	return

}
