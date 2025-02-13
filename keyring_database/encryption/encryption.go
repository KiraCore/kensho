package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	SALT_SIZE  = 16
	NONCE_SIZE = 12
	KEY_SIZE   = 32
	ITER_COUNT = 100000 // Increase for better security
)

func DeriveKey(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, ITER_COUNT, KEY_SIZE, sha256.New)
}

func Encrypt(dataToEncrypt string, password string) (string, error) {
	salt := make([]byte, SALT_SIZE)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	key := DeriveKey(password, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, NONCE_SIZE)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nil, nonce, []byte(dataToEncrypt), nil)

	finalData := append(salt, nonce...)
	finalData = append(finalData, ciphertext...)

	return base64.StdEncoding.EncodeToString(finalData), nil
}

func Decrypt(encryptedData string, password string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}

	if len(data) < SALT_SIZE+NONCE_SIZE {
		return "", fmt.Errorf("invalid encrypted data")
	}
	salt := data[:SALT_SIZE]
	nonce := data[SALT_SIZE : SALT_SIZE+NONCE_SIZE]
	ciphertext := data[SALT_SIZE+NONCE_SIZE:]

	key := DeriveKey(password, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
