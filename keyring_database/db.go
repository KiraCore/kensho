package ipdatabase

import (
	"net"

	ipCredentials "github.com/KiraCore/kensho/keyring_database/host_credentials"
	hostRegistry "github.com/KiraCore/kensho/keyring_database/host_registry"
)

type IP_DB struct {
	hostReg   *hostRegistry.HostRegistry
	hostCreds *ipCredentials.CredentialsManager
}

func NewIPDataBase() (*IP_DB, error) {
	ipreg, err := hostRegistry.NewIPRegistry("./test")
	if err != nil {
		return nil, err
	}
	creds := ipCredentials.NewKeyringManager()
	return &IP_DB{hostReg: ipreg, hostCreds: creds}, nil

}

// ip - host's ip
// port - host's port
// user - host's user
// secret - password/key path
// key - if secret is a key or path
func (db IP_DB) Add(ip, port, user, secret string, key bool) error {
	id := net.JoinHostPort(ip, port)
	err := db.hostCreds.AddCredentials(id, ipCredentials.Credentials{
		Key:    key,
		User:   user,
		Secret: secret,
	})
	if err != nil {
		return err
	}
	err = db.hostReg.AddIP(id)
	if err != nil {
		return err
	}
	return nil
}

func (db IP_DB) Remove(id string) error {
	err := db.hostReg.DeleteIP(id)
	if err != nil {
		return err
	}
	err = db.hostCreds.RemoveCredentials(id)
	if err != nil {
		return err
	}

	return nil
}

func (db IP_DB) Get(id string) (*ipCredentials.Credentials, error) {
	creds, err := db.hostCreds.GetCredentials(id)
	if err != nil {
		return nil, err
	}
	return creds, nil
}

func (db IP_DB) GetAll() []string {
	return db.hostReg.ListIPs()
}
