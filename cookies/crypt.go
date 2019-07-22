package cookies

import (
	"crypto/aes"
	"crypto/cipher"
)

var IV = []byte("1234567812345678")

func (c Crypt) Crypt(data []byte) []byte {
	stream := cipher.NewCTR(c.Cipher, IV)
	stream.XORKeyStream(data, data)
	return data
}

func GetCryptInstance(key string) Crypt {
	c, _ := aes.NewCipher([]byte(key))
	return Crypt{Cipher: c}
}
