package encryptions

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

var encryptionKey []byte

// InitEncryptionKey loads the encryption key from an environment variable
func InitEncryptionKey() error {
	key := []byte(os.Getenv("ENCRYPTION_KEY"))
	// Validate key length (AES requires 16, 24, or 32 bytes)
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return errors.New("invalid ENCRYPTION_KEY length: must be 16, 24, or 32 bytes")
	}

	encryptionKey = key
	return nil
}

// Encrypt encrypts a string using AES CFB mode
func Encrypt(data string) (string, error) {
	if encryptionKey == nil {
		return "", errors.New("encryption key is not initialized, call InitEncryptionKey() first")
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	plainText := []byte(data)
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// Decrypt decrypts a Base64-encoded AES CFB ciphertext
func Decrypt(data string) (string, error) {
	if encryptionKey == nil {
		return "", errors.New("encryption key is not initialized, call InitEncryptionKey() first")
	}

	cipherText, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	if len(cipherText) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return string(cipherText), nil
}
