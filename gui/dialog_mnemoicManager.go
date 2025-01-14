package gui

import (
	"fmt"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	dialogWizard "github.com/KiraCore/kensho/gui/dialogs"
	mnemonicHelper "github.com/KiraCore/kensho/helper/mnemonicHelper"
	vlg "github.com/KiraCore/tools/validator-key-gen/MnemonicsGenerator"

	"github.com/atotto/clipboard"
)

func showMnemonicManagerDialog(g *Gui, mnemonicBinding binding.String, doneAction binding.DataListener) {
	var wizard *dialogWizard.Wizard
	mnemonicDisplay := container.NewGridWithColumns(2)
	localMnemonicBinding := binding.NewString()
	oldM, _ := mnemonicBinding.Get()
	localMnemonicBinding.Set(oldM)
	warningConfirmDataListener := binding.NewDataListener(func() {
		lMnemonic, err := localMnemonicBinding.Get()
		if err != nil {
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
		}
		err = mnemonicBinding.Set(lMnemonic)
		if err != nil {
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
		}
		doneAction.DataChanged()
		wizard.Hide()
	})
	warningMessage := `By clicking "Proceed," you confirm that you have saved your mnemonic. You will no longer be able to see your mnemonic a second time. Make sure you have securely stored it before proceeding.
If you have not please press "Return" and save your mnemonic.`
	doneButton := widget.NewButton("Done", func() {
		showWarningMessageWithConfirmation(g, warningMessage, warningConfirmDataListener)
	})
	doneButton.Disable()
	var content *fyne.Container

	showDetailsButton := widget.NewButton("Show Details", func() {
		showMasterMnemonicDetails(g, localMnemonicBinding)
	})
	showDetailsButton.Disable()
	// doing this to display mnemonic if it was already generated
	m, err := mnemonicBinding.Get()
	if err != nil {
		g.showErrorDialog(err, binding.NewDataListener(func() {}))
		return
	}
	if m != "" {
		mnemonicWords := strings.Split(m, " ")
		mnemonicDisplay.RemoveAll()
		for i, w := range mnemonicWords {
			mnemonicDisplay.Add(widget.NewLabel(fmt.Sprintf("%v. %v", i+1, w)))
		}
	}
	//

	mnemonicChanged := binding.NewDataListener(func() {
		m, err := localMnemonicBinding.Get()
		if err != nil {
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
			return
		}

		err = mnemonicHelper.ValidateMnemonic(m)
		if err != nil {
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
			return
		}
		doneButton.Enable()
		showDetailsButton.Enable()
		mnemonicWords := strings.Split(m, " ")
		mnemonicDisplay.RemoveAll()
		for i, w := range mnemonicWords {
			mnemonicDisplay.Add(widget.NewLabel(fmt.Sprintf("%v. %v", i+1, w)))
		}
		content.Refresh()
	})

	closeButton := widget.NewButton("Close", func() {
		wizard.Hide()
	})

	doneEnteringMnemonicListener := binding.NewDataListener(func() {
		mnemonicChanged.DataChanged()
	})
	enterMnemonicManuallyButton := widget.NewButton("Enter your mnemonic", func() {

		showMnemonicEntryDialog(g, localMnemonicBinding, doneEnteringMnemonicListener)
	})

	copyButton := widget.NewButtonWithIcon("Copy", theme.FileIcon(), func() {
		data, _ := localMnemonicBinding.Get()
		err = clipboard.WriteAll(data)
		if err != nil {
			log.Println(err)
			return
		}
	})

	generateButton := widget.NewButton("Generate", func() {
		masterMnemonic, err := mnemonicHelper.GenerateMnemonic()
		if err != nil {
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
			return
		}

		err = mnemonicHelper.ValidateMnemonic(masterMnemonic.String())
		if err != nil {
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
			return
		}
		err = localMnemonicBinding.Set(masterMnemonic.String())
		if err != nil {
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
			return
		}

		mnemonicChanged.DataChanged()
	})

	content = container.NewBorder(
		nil,
		container.NewVBox(enterMnemonicManuallyButton, container.NewVBox(container.NewGridWithColumns(2, generateButton, copyButton)), showDetailsButton, closeButton, doneButton),
		nil,
		nil,
		mnemonicDisplay,
	)

	wizard = dialogWizard.NewWizard("Mnemonic setup", content)
	wizard.Show(g.Window)
	wizard.Resize(fyne.NewSize(400, 700))
}

func showMnemonicEntryDialog(g *Gui, mnemonicBinding binding.String, doneAction binding.DataListener) {
	var wizard *dialogWizard.Wizard
	infoLabel := widget.NewLabel("Enter your mnemonic")
	infoLabel.Wrapping = fyne.TextWrapWord
	mnemonicEntry := widget.NewEntry()
	mnemonicEntry.Wrapping = fyne.TextWrapWord
	mnemonicEntry.MultiLine = true
	closeButton := widget.NewButton("Close", func() {
		wizard.Hide()
	})

	doneButton := widget.NewButton("Done", func() {
		mnemonicBinding.Set(mnemonicEntry.Text)
		doneAction.DataChanged()
		wizard.Hide()
	})
	doneButton.Disable()

	mnemonicEntry.OnChanged = func(s string) {
		err := mnemonicHelper.ValidateMnemonic(mnemonicEntry.Text)
		if err != nil {
			infoLabel.SetText("Mnemonic is not valid")
			doneButton.Disable()
		} else {
			infoLabel.SetText("Mnemonic is valid")
			doneButton.Enable()
		}
	}

	content := container.NewBorder(
		infoLabel,
		container.NewVBox(closeButton, doneButton),
		nil,
		nil,
		(mnemonicEntry),
	)

	wizard = dialogWizard.NewWizard("Mnemonic setup", content)
	wizard.Show(g.Window)
	wizard.Resize(fyne.NewSize(900, 200))
}

func showMasterMnemonicDetails(g *Gui, mnemonicBinding binding.String) {
	var wizard *dialogWizard.Wizard

	closeButton := widget.NewButton("Close", func() {
		wizard.Hide()
	})

	mstrMnmc, _ := mnemonicBinding.Get()
	mnemonicSet, err := vlg.MasterKeysGen([]byte(mstrMnmc), vlg.DefaultPrefix, vlg.DefaultPath, "")
	if err != nil {
		g.showErrorDialog(err, binding.NewDataListener(func() {}))
	}
	mnemonicsData := binding.NewString()
	kiraAddress, err := mnemonicHelper.GetKiraAddressFromMnemonic(mnemonicSet.ValidatorAddrMnemonic)
	if err != nil {
		g.showErrorDialog(err, binding.NewDataListener(func() {}))
	}
	mnemonicsData.Set(fmt.Sprintf("VALIDATOR_ADDR_MNEMONIC=%s\n\nVALIDATOR_NODE_MNEMONIC=%s\n\nVALIDATOR_VAL_MNEMONIC=%s\n\nSIGNER_ADDR_MNEMONIC=%s\n\n\nVALIDATOR_ADDRESS=%s\nVALIDATOR_NODE_ID=%s", string(mnemonicSet.ValidatorAddrMnemonic), string(mnemonicSet.ValidatorNodeMnemonic), string(mnemonicSet.ValidatorValMnemonic), string(mnemonicSet.SignerAddrMnemonic), kiraAddress, mnemonicSet.ValidatorNodeId))

	copyButton := widget.NewButtonWithIcon("Copy", theme.FileIcon(), func() {
		data, _ := mnemonicsData.Get()
		err = clipboard.WriteAll(data)
		if err != nil {
			log.Println(err)
			return
		}
	})
	infoLabel := widget.NewLabelWithData(mnemonicsData)
	infoLabel.Wrapping = fyne.TextWrapWord
	infoContent := container.NewVScroll(
		infoLabel,
	)

	content := container.NewBorder(
		nil,
		container.NewGridWithColumns(3, copyButton, layout.NewSpacer(), closeButton),
		nil,
		nil,
		infoContent,
	)

	wizard = dialogWizard.NewWizard("Mnemonic details", content)
	wizard.Show(g.Window)
	wizard.Resize(fyne.NewSize(900, 400))
}
