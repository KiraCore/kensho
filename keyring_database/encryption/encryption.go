package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/KiraCore/kensho/types"
	"github.com/zalando/go-keyring"
)

const username = "KenshoEncryptionKey"
const fallbackKeyFile = "encryption_key.txt"

func Decrypt(encryptedText string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}
	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return string(ciphertext), nil
}

func Encrypt(text string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println("errror creating new cipher block ")
		return "", err
	}
	plaintext := []byte(text)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func getEncryptionKey(homeDir string) ([]byte, error) {
	key, err := keyring.Get(types.APP_NAME, username)
	if err == keyring.ErrNotFound {
		fmt.Println("Key not found in system keyring. Falling back to file storage.")
		return getEncryptionKeyFromFile(homeDir)
	} else if err != nil {
		fmt.Println("Keyring error. Falling back to file storage.")
		return getEncryptionKeyFromFile(homeDir)
	}

	return base64.StdEncoding.DecodeString(key)
}

func getEncryptionKeyFromFile(homeDir string) ([]byte, error) {
	var keyPath string
	if homeDir == "" {
		keyPath = filepath.Join(os.TempDir(), fallbackKeyFile)

	} else {
		keyPath = filepath.Join(homeDir, fallbackKeyFile)
	}

	if _, err := os.Stat(keyPath); errors.Is(err, os.ErrNotExist) {
		newKey := make([]byte, 32)
		_, err := rand.Read(newKey)
		if err != nil {
			return nil, fmt.Errorf("failed to generate encryption key: %v", err)
		}

		encodedKey := base64.StdEncoding.EncodeToString(newKey)
		if err := os.WriteFile(keyPath, []byte(encodedKey), 0600); err != nil {
			return nil, fmt.Errorf("failed to write key to file: %v", err)
		}
		return newKey, nil
	}

	data, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key from file: %v", err)
	}

	return base64.StdEncoding.DecodeString(string(data))
}
