package mnemonichelper

import (
	"fmt"
	"log"

	cosmosBIP39 "github.com/cosmos/go-bip39"
	kiraMnemonicGen "github.com/kiracore/tools/bip39gen/cmd"
	"github.com/kiracore/tools/bip39gen/pkg/bip39"
)

func GenerateMnemonic() (masterMnemonic bip39.Mnemonic, err error) {
	log.Println("generating new mnemonic")
	masterMnemonic = kiraMnemonicGen.NewMnemonic()
	masterMnemonic.SetRandomEntropy(24)
	masterMnemonic.Generate()

	return masterMnemonic, nil
}

func ValidateMnemonic(mnemonic string) error {
	check := cosmosBIP39.IsMnemonicValid(mnemonic)
	if !check {
		return fmt.Errorf("mnemonic <%v> is not valid", mnemonic)
	}
	return nil
}

func GetKiraAddressFromMnemonic(mnemonic []byte) (string, error) {

	return "someAddress", nil
}
