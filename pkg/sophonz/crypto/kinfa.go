// Package crypto provides the KINFA AES-256-CBC helpers used to decrypt
// (and encrypt) user identifiers carried on telemetry.
//
// The key is process-global and must be initialized once via
// InitializeKey before Encrypt/Decrypt are called (typically from the
// processor factory using the AES_DECRYPT_KEY_KINFA environment variable).
// A fixed zero IV is used to match the upstream KINFA scheme.
package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

var (
	// key is the AES-256 key (32 bytes).
	key []byte
	// iv is a 16-byte zero IV, matching the KINFA scheme.
	iv = make([]byte, aes.BlockSize)
)

// InitializeKey sets the process-global AES-256 key used by Encrypt/Decrypt.
func InitializeKey(decryptKey string) {
	key = []byte(decryptKey)
}

func pkcs5Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padText...)
}

func pkcs5UnPadding(src []byte) []byte {
	length := len(src)
	if length == 0 {
		return nil
	}
	unpadding := int(src[length-1])
	if unpadding <= 0 || unpadding > length {
		return nil
	}
	return src[:(length - unpadding)]
}

// Encrypt AES-256-CBC encrypts plainText and returns a base64 string.
func Encrypt(plainText string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	padded := pkcs5Padding([]byte(plainText), block.BlockSize())

	mode := cipher.NewCBCEncrypter(block, iv)
	encrypted := make([]byte, len(padded))
	mode.CryptBlocks(encrypted, padded)

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// Decrypt base64-decodes and AES-256-CBC decrypts cipherTextBase64.
func Decrypt(cipherTextBase64 string) (string, error) {
	cipherText, err := base64.StdEncoding.DecodeString(cipherTextBase64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(cipherText)%block.BlockSize() != 0 {
		return "", fmt.Errorf("crypto/cipher: input not full blocks - ciphertext length %d is not a multiple of block size %d", len(cipherText), block.BlockSize())
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(cipherText))
	mode.CryptBlocks(decrypted, cipherText)

	return string(pkcs5UnPadding(decrypted)), nil
}
