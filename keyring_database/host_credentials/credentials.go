package host_credentials

import (
	"encoding/json"

	"github.com/KiraCore/kensho/keyring_database/encryption"
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
	encryptedData, err := encryption.Encrypt(string(data), []byte("password"))
	if err != nil {
		return err
	}
	err = keyring.Set(types.APP_NAME, id, encryptedData)
	if err != nil {
		return err
	}
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
	encryptedData, err := keyring.Get(types.APP_NAME, id)
	if err != nil {
		return nil, err
	}
	decryptedData, err := encryption.Decrypt(encryptedData, []byte("password"))
	if err != nil {
		return nil, err
	}
	var cred Credentials
	err = json.Unmarshal([]byte(decryptedData), &cred)
	if err != nil {
		return nil, err
	}
	return &cred, nil
}
