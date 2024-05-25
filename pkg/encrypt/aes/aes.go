package aes

import (
	"bytes"
	aes "crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"strings"
)

var cipherKey []byte

func InitAes(s string) {
	cipherKey = []byte(s)
}

func addBase64Padding(value string) string {
	m := len(value) % 4
	if m != 0 {
		value += strings.Repeat("=", 4-m)
	}

	return value
}

func removeBase64Padding(value string) string {
	return strings.Replace(value, "=", "", -1)
}

func pad(src []byte) []byte {
	padding := aes.BlockSize - (len(src) % aes.BlockSize)
	padtext := []byte(bytes.Repeat([]byte{byte(padding)}, padding))
	return append(src, padtext...)
}

func unpad(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])

	if unpadding > length {
		return nil, errors.New("unpad error. This could happen when incorrect encryption key is used")
	}

	return src[:length-unpadding], nil
}

func Encrypt(text string) (string, error) {
	block, err := aes.NewCipher(cipherKey)
	if err != nil {
		return "", err
	}

	msg := pad([]byte(text))
	cipherText := make([]byte, len(msg)+aes.BlockSize)
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(cipherText[aes.BlockSize:], []byte(msg))
	finalMsg := removeBase64Padding(base64.URLEncoding.EncodeToString(cipherText))

	return finalMsg, nil
}

func Decrypt(text string) (string, error) {
	block, err := aes.NewCipher(cipherKey)
	if err != nil {
		return "", err
	}

	decodeMsg, err := base64.URLEncoding.DecodeString(addBase64Padding(text))
	if err != nil {
		return "", err
	}

	if (len(decodeMsg) % aes.BlockSize) != 0 {
		return "", errors.New("blocksize must be multiple of decode message")
	}

	iv := decodeMsg[:aes.BlockSize]
	msg := decodeMsg[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(msg, msg)

	unpadMsg, err := unpad(msg)
	if err != nil {
		return "", err
	}

	return string(unpadMsg), nil
}
