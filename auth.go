package cosmosdb

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

func authorize(str, key string) (string, error) {
	var ret string
	enc := base64.StdEncoding
	salt, err := enc.DecodeString(key)
	if err != nil {
		return ret, err
	}
	hmac := hmac.New(sha256.New, salt)
	hmac.Write([]byte(str))
	b := hmac.Sum(nil)

	ret = enc.EncodeToString(b)
	return ret, nil
}
