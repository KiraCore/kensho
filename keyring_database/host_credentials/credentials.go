package host_credentials

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"

	"github.com/KiraCore/kensho/types"
	"github.com/zalando/go-keyring"
)

type Credentials struct {
	Key    bool   `json:"key"`
	User   string `json:"user"`
	Secret string `json:"secret"`
}

func NewKeyringManager() *CredentialsManager {
	return &CredentialsManager{}
}

type CredentialsManager struct{}

func (CredentialsManager) AddCredentials(id string, credentials Credentials) error {
	data, err := json.Marshal(credentials)
	if err != nil {
		return err
	}
	err = keyring.Set(types.APP_NAME, id, string(data))
	if err != nil {
		return err
	}
	// fmt.Printf("%+v", keyring.)
	return nil
}

func (CredentialsManager) RemoveCredentials(id string) error {
	err := keyring.Delete(types.APP_NAME, id)
	if err != nil {
		return err
	}
	return nil
}

func (CredentialsManager) GetCredentials(id string) (*Credentials, error) {
	out, err := keyring.Get(types.APP_NAME, id)
	if err != nil {
		return nil, err
	}

	var cred Credentials
	err = json.Unmarshal([]byte(out), &cred)
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

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
