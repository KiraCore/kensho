package gui

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
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	dialogWizard "github.com/KiraCore/kensho/gui/dialogs"
	"github.com/KiraCore/kensho/helper/gssh"
	"github.com/KiraCore/kensho/types"
	"github.com/fyne-io/terminal"
	"github.com/zalando/go-keyring"
	"golang.org/x/crypto/ssh"
)

const username = "KenshoEncryptionKey"
const fallbackKeyFile = "encryption_key.txt"

func (g *Gui) ShowConnect() {

	var wizard *dialogWizard.Wizard

	//join to new host tab
	join := func() *fyne.Container {
		encryptionKey, err := getEncryptionKey()
		if err != nil {
			fmt.Println("Error getting encryption key:", err)
			return nil
		}

		savedIp := g.Preferences.String("ip")
		savedPort := g.Preferences.String("port")
		savedUsername := g.Preferences.String("username")
		savedPkCheckbox := g.Preferences.Bool("pkc")
		savedPkPath := g.Preferences.String("pkpath")
		savedSaveCheckbox := g.Preferences.Bool("svc")

		encryptedPassword := g.Preferences.String("password")

		var savedPassword string
		if encryptedPassword != "" {
			savedPassword, _ = decrypt(encryptedPassword, encryptionKey)
		}

		userEntry := widget.NewEntry()
		ipEntry := widget.NewEntry()
		portEntry := widget.NewEntry()
		passwordEntry := widget.NewPasswordEntry()
		errorLabel := widget.NewLabel("")
		keyPathEntry := widget.NewEntry()
		rawKeyEntry := widget.NewEntry()
		passphraseEntry := widget.NewPasswordEntry()
		passphraseEntry.Hide()
		var privKeyState bool
		var passphraseState bool
		var rawKeyState bool
		var saveState bool

		userEntry.SetText(savedUsername)
		ipEntry.SetText(savedIp)
		portEntry.SetText(savedPort)
		portEntry.PlaceHolder = "22"
		passwordEntry.SetText(savedPassword)
		keyPathEntry.SetText(savedPkPath)

		passphraseEntry.SetText(savedPassword)
		passphraseEntry.Validator = func(s string) error {
			if s == "" {
				return fmt.Errorf("enter your passphrase")
			}
			return nil
		}
		addressBoxEntry := container.NewBorder(nil, nil, nil, container.NewHBox(widget.NewLabel(":"), portEntry), ipEntry)
		rawKeyEntry.PlaceHolder = "private key in plain text"
		rawKeyEntry.MultiLine = true
		rawKeyEntry.Wrapping = fyne.TextWrapBreak
		keyPathEntry.PlaceHolder = "path to your private key"
		passphraseEntry.PlaceHolder = "your passphrase"
		passphraseCheck := widget.NewCheck("SSH passphrase key", func(b bool) {
			passphraseState = b
			if passphraseState {
				passphraseEntry.Show()
			} else {
				passphraseEntry.Hide()
			}
		})

		rawKeyCheck := widget.NewCheck("Raw key", func(b bool) {})
		rawKeyEntry.OnChanged = func(s string) {

			check, err := gssh.CheckIfPassphraseNeeded([]byte(s))
			if err != nil {
				return
			}
			if check {
				passphraseCheck.SetChecked(true)
			} else {
				passphraseCheck.SetChecked(false)
			}
		}
		keyPathEntry.OnChanged = func(s string) {
			b, err := os.ReadFile(s)
			if err != nil {
				return
			}
			check, err := gssh.CheckIfPassphraseNeeded(b)
			if err != nil {
				return
			}
			if check {
				passphraseCheck.SetChecked(true)
			} else {
				passphraseCheck.SetChecked(false)
			}
		}

		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if reader == nil {
				return
			}

			uri := reader.URI().Path()
			keyPathEntry.SetText(uri)
			log.Println("Opened file: ", uri)

		}, g.Window)

		openFileDialogButton := widget.NewButtonWithIcon("", theme.FileIcon(), func() { fileDialog.Show() })
		privKeyBox := container.NewBorder(
			widget.NewLabel("Select private key"),
			nil, nil,
			openFileDialogButton,
			keyPathEntry,
		)

		rawPrivKeyBox := container.NewBorder(
			widget.NewLabel("Enter your private key"),
			nil, nil,
			nil,
			rawKeyEntry,
		)

		filePrivKeyBox := container.NewVBox(
			privKeyBox,
		)

		passwordBoxEntry := container.NewVBox(
			widget.NewLabel("Password"),
			passwordEntry,
		)
		keyEntryBox := container.NewStack(passwordBoxEntry)

		privKeyBoxEntry := container.NewVBox(
			filePrivKeyBox,
			passphraseEntry,
			container.NewHBox(passphraseCheck, rawKeyCheck),
		)

		privKeyCheck := widget.NewCheck("Join with private key", func(b bool) {
			privKeyState = b
			if b {
				keyEntryBox.Objects = []fyne.CanvasObject{privKeyBoxEntry}
			} else {
				keyEntryBox.Objects = []fyne.CanvasObject{passwordBoxEntry}
			}
		})

		privKeyCheck.SetChecked(savedPkCheckbox)

		saveCheck := widget.NewCheck("Remember credentials", func(b bool) {
			saveState = b
		})

		saveCheck.SetChecked(savedSaveCheckbox)

		rawKeyCheck.OnChanged = func(b bool) {
			rawKeyState = b
			if b {
				privKeyBoxEntry.Objects[0] = rawPrivKeyBox
			} else {
				privKeyBoxEntry.Objects[0] = filePrivKeyBox
			}
		}

		errorLabel.Wrapping = 2

		submitFunc := func() {
			g.WaitDialog.ShowWaitDialog()
			var err error
			ip := strings.TrimSpace(ipEntry.Text)
			port := ""
			if portEntry.Text == "" {
				port = "22"
			} else {
				port = strings.TrimSpace(portEntry.Text)
			}
			address := fmt.Sprintf("%v:%v", ip, (port))

			if saveState {
				encryptedPassword, _ := encrypt(passwordEntry.Text, encryptionKey)
				g.Preferences.SetString("ip", ipEntry.Text)
				g.Preferences.SetString("port", portEntry.Text)
				g.Preferences.SetString("username", userEntry.Text)
				g.Preferences.SetString("password", encryptedPassword)
				g.Preferences.SetBool("pkc", privKeyState)
				g.Preferences.SetString("pkpath", keyPathEntry.Text)
				g.Preferences.SetBool("svc", saveState)
			} else {
				g.Preferences.SetString("ip", "")
				g.Preferences.SetString("port", "")
				g.Preferences.SetString("username", "")
				g.Preferences.SetString("password", "")
				g.Preferences.SetBool("pkc", false)
				g.Preferences.SetString("pkpath", "")
				g.Preferences.SetBool("svc", saveState)
			}

			if privKeyState {
				var b []byte
				var c *ssh.Client

				g.sshClient, err = func() (*ssh.Client, error) {
					log.Println("Raw key state: ", rawKeyState)
					if rawKeyState {
						b = []byte(rawKeyEntry.Text)
					} else {
						b, err = os.ReadFile(keyPathEntry.Text)
						if err != nil {
							return nil, err
						}
					}

					check, err := gssh.CheckIfPassphraseNeeded(b)
					log.Println("Passphrase check:", check)
					if err != nil {
						log.Printf("Error when checking if key need a passphrase: %v", err.Error())
						return nil, err
					}
					if check {
						if passphraseEntry.Hidden {
							passphraseEntry.Validate()
							passphraseEntry.SetValidationError(fmt.Errorf("passphrase required"))
							passphraseCheck.SetChecked(true)
						}

						c, err = gssh.MakeSSH_ClientWithPrivKeyAndPassphrase(address, userEntry.Text, b, []byte(passphraseEntry.Text))
						if err != nil {
							log.Printf("error when creating ssh client: %v", err.Error())
							return nil, err
						}
					} else {
						if !passphraseEntry.Hidden {
							passphraseCheck.SetChecked(false)

						}
						c, err = gssh.MakeSSH_ClientWithPrivKey(address, userEntry.Text, b)
						if err != nil {
							return nil, err
						}
					}
					return c, nil
				}()
			} else {
				g.sshClient, err = gssh.MakeSHH_ClientWithPassword(address, userEntry.Text, passwordEntry.Text)
			}
			if err != nil {
				log.Println("ERROR submitting:", err.Error())
				errorLabel.SetText(fmt.Sprintf("ERROR: %s", err.Error()))
				g.showErrorDialog(err, binding.NewDataListener(func() {}))
			} else {
				err := TryToRunSSHSessionForTerminal(g)
				if err != nil {
					g.showErrorDialog(fmt.Errorf("unable to create terminal instance, disabling terminal: %v", err.Error()), binding.NewDataListener(func() {}))
					g.Terminal.Term = terminal.New()

				}
				g.Host = &Host{
					IP: ip,
				}
				if !privKeyState {
					g.Host.UserPassword = &passwordEntry.Text
				}
				go g.sshAliveTracker()
				g.ConnectionStatusBinding.Set(true)
				wizard.Hide()
			}
			defer g.WaitDialog.HideWaitDialog()
		}

		// / test ui block
		testButton := widget.NewButton("connect to tested env", func() {
			ipEntry.Text = "192.168.1.101"
			userEntry.Text = "d"
			passwordEntry.Text = "d"
			passphraseCheck.SetChecked(false)

			submitFunc()
		})

		if !g.DeveloperMode {
			testButton.Disable()
			testButton.Hide()
		}

		///

		ipEntry.OnSubmitted = func(s string) { submitFunc() }
		userEntry.OnSubmitted = func(s string) { submitFunc() }
		passwordEntry.OnSubmitted = func(s string) { submitFunc() }
		connectButton := widget.NewButton("Connect to remote host", func() { submitFunc() })

		logging := container.NewBorder(
			container.NewVBox(
				widget.NewLabel("IP and Port"),
				addressBoxEntry,
				widget.NewLabel("User"),
				userEntry,
				keyEntryBox,
				privKeyCheck,
				saveCheck,
				connectButton,
				testButton,
			),
			nil, nil, nil,
			container.NewBorder(nil, nil, nil, nil, container.NewVScroll(errorLabel)),
		)
		return logging
	}

	wizard = dialogWizard.NewWizard("Create ssh connection", join())
	wizard.Show(g.Window)
	wizard.Resize(fyne.NewSize(350, 540))
}

func (g *Gui) sshAliveTracker() {
	errorDoneBinding := binding.NewDataListener(func() {
		// g.ShowConnect()
	})
	g.ConnectionCount++

	err := g.sshClient.Wait()
	if err != nil {
		log.Printf("SSH was interrupted: %v", err.Error())
		g.ConnectionStatusBinding.Set(false)

		g.showErrorDialog(fmt.Errorf("SSH connection was disconnected, reason: %v", err.Error()), errorDoneBinding)
	}

}

func getEncryptionKey() ([]byte, error) {
	key, err := keyring.Get(types.APP_NAME, username)
	if err == keyring.ErrNotFound {
		fmt.Println("Key not found in system keyring. Falling back to file storage.")
		return getEncryptionKeyFromFile()
	} else if err != nil {
		fmt.Println("Keyring error. Falling back to file storage.")
		return getEncryptionKeyFromFile()
	}

	return base64.StdEncoding.DecodeString(key)
}

func getEncryptionKeyFromFile() ([]byte, error) {
	keyPath := filepath.Join(os.TempDir(), fallbackKeyFile)

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

func decrypt(encryptedText string, key []byte) (string, error) {
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

func encrypt(text string, key []byte) (string, error) {
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
