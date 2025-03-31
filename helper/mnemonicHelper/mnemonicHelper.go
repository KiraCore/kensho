package mnemonichelper

import (
	"fmt"
	"log"

	"github.com/cosmos/cosmos-sdk/codec"

	vlg "github.com/KiraCore/tools/validator-key-gen/MnemonicsGenerator"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	ltypes "github.com/KiraCore/kensho/types"
	ctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"
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
	bAddress, err := ConvertMnemonicToAddrBytes(string(mnemonic), vlg.DefaultPath)
	if err != nil {
		return "", err
	}

	addr, err := ConvertRawBytesAddressToKira(bAddress)
	if err != nil {
		return "", err
	}
	return addr, nil
}

func ConvertMnemonicToAddrBytes(mnemonic, hdPath string) ([]byte, error) {
	interfaceRegistry := ctypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	kr := keyring.NewInMemory(cdc)
	key, err := kr.NewAccount("tmp", mnemonic, "", hdPath, hd.Secp256k1)
	if err != nil {
		return nil, err
	}
	addr, err := key.GetAddress()
	if err != nil {
		return nil, err
	}
	return addr.Bytes(), nil
}

func ConvertRawBytesAddressToKira(addrBytes []byte) (string, error) {
	prefix := ltypes.KIRA_ADDRESS_PREFIX
	newBech32Addr, err := types.Bech32ifyAddressBytes(prefix, addrBytes)
	if err != nil {
		fmt.Printf("Error converting to Bech32 address with new prefix: %s\n", err)
		return "", err
	}
	return newBech32Addr, nil
}
