package host_credentials

import (
	"encoding/json"

	"github.com/zalando/go-keyring"
)

type Credentials struct {
	Key    bool   `json:"key"`
	User   string `json:"user"`
	Secret string `json:"secret"`
}

const Service string = "kensho"

func NewKeyringManager() *CredentialsManager {
	return &CredentialsManager{}
}

type CredentialsManager struct{}

func (CredentialsManager) AddCredentials(id string, credentials Credentials) error {
	data, err := json.Marshal(credentials)
	if err != nil {
		return err
	}
	err = keyring.Set(Service, id, string(data))
	if err != nil {
		return err
	}
	// fmt.Printf("%+v", keyring.)
	return nil
}

func (CredentialsManager) RemoveCredentials(id string) error {
	err := keyring.Delete(Service, id)
	if err != nil {
		return err
	}
	return nil
}

func (CredentialsManager) GetCredentials(id string) (*Credentials, error) {
	out, err := keyring.Get(Service, id)
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
