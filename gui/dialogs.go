package gui

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	dialogWizard "github.com/KiraCore/kensho/gui/dialogs"
	"github.com/KiraCore/kensho/helper/gssh"
	"github.com/KiraCore/kensho/utils"
	"github.com/atotto/clipboard"
)

type WaitDialog struct {
	wizard *dialogWizard.Wizard
	g      *Gui
}

// Wait dialog init
func NewWaitDialog(g *Gui) *WaitDialog {
	var wizard *dialogWizard.Wizard

	loadingWidget := widget.NewProgressBarInfinite()
	content := container.NewVBox(loadingWidget)

	wizard = dialogWizard.NewWizard("Loading", content)
	return &WaitDialog{
		wizard: wizard,
		g:      g,
	}
}

func (w *WaitDialog) ShowWaitDialog() {
	w.wizard.Show(w.g.Window)
}

func (w *WaitDialog) HideWaitDialog() {
	w.wizard.Hide()
}

func (g *Gui) showErrorDialog(err error, closeListener binding.DataListener) {
	var wizard *dialogWizard.Wizard
	log.Println("Enabling error dialog with err:", err.Error())
	errorLabel := widget.NewLabel(err.Error())
	errorLabel.Wrapping = fyne.TextWrapWord

	mainDialogScreen :=
		container.NewBorder(nil, container.NewVBox(

			widget.NewButton("Copy", func() {
				err = clipboard.WriteAll(errorLabel.Text)
				if err != nil {
					return
				}
			}),
			widget.NewButton("Close", func() { wizard.Hide(); closeListener.DataChanged() }),
		), nil, nil,

			container.NewVScroll(errorLabel),
		)
	wizard = dialogWizard.NewWizard("Error", mainDialogScreen)
	wizard.Show(g.Window)
	wizard.Resize(fyne.NewSize(400, 400))

}

func showInfoDialog(g *Gui, infoTitle, infoString string) {
	var wizard *dialogWizard.Wizard
	closeButton := widget.NewButton("Close", func() { wizard.Hide() })
	infoLabel := widget.NewLabel(infoString)
	infoLabel.Wrapping = 2
	copyButton := widget.NewButton("Copy", func() { utils.CopyToClipboard(infoLabel.Text) })
	content := container.NewBorder(nil, container.NewVBox(copyButton, closeButton), nil, nil,
		container.NewHScroll(
			container.NewVScroll(
				infoLabel,
			),
		),
	)

	wizard = dialogWizard.NewWizard(infoTitle, content)
	wizard.Show(g.Window)
	wizard.Resize(fyne.NewSize(400, 400))
}

func showCmdExecDialogAndRunCmdV4(g *Gui, infoMSG string, cmd string, autoHideCheck bool, errorBinding binding.Bool, errorMessageBinding binding.String) {
	outputChannel := make(chan string)
	errorChannel := make(chan gssh.ResultV2)
	// go gssh.ExecuteSSHCommandV2(g.sshClient, cmd, outputChannel, errorChannel)
	go gssh.ExecuteSSHCommandV3(g.sshClient, cmd, outputChannel, errorChannel)

	var wizard *dialogWizard.Wizard
	outputMsg := binding.NewString()
	statusMsg := binding.NewString()
	statusMsg.Set("Loading...")
	loadingWidget := widget.NewProgressBarInfinite()

	label := widget.NewLabelWithData(outputMsg)
	label.Wrapping = fyne.TextWrapWord

	closeButton := widget.NewButton("Done", func() { wizard.Hide() })
	outputScroll := container.NewVScroll(label)

	loadingDialog := container.NewBorder(
		widget.NewLabelWithData(statusMsg),
		container.NewVBox(loadingWidget, closeButton),
		nil,
		nil,
		container.NewHScroll(outputScroll),
	)
	closeButton.Hide()
	wizard = dialogWizard.NewWizard(infoMSG, loadingDialog)
	wizard.Show(g.Window)
	wizard.Resize(fyne.NewSize(300, 400))
	if autoHideCheck {
		defer wizard.Hide()
	}
	var out string
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for line := range outputChannel {
			cleanLine := cleanString(line)
			out = fmt.Sprintf("%s\n%s", out, cleanLine)
			outputMsg.Set(out)
			outputScroll.ScrollToBottom()
		}
	}()
	var errcheck gssh.ResultV2
	go func() {
		defer wg.Done()

		errcheck = <-errorChannel
	}()

	wg.Wait()

	loadingWidget.Hide()
	closeButton.Show()
	if errcheck.Err != nil {
		log.Printf("Unable to execute executing: <%v>, error: %v, %v ", cmd, errcheck.Err.Error(), string(out))
		errorBinding.Set(true)
		errorMessageBinding.Set(fmt.Sprintf("Out: %v, Error: %v", string(out), errcheck.Err.Error()))
		statusMsg.Set(fmt.Sprintf("Error:\n%s", errcheck.Err))
	} else {
		errorBinding.Set(false)
		wizard.ChangeTitle("Done")
		statusMsg.Set("Successes")
	}
	outputScroll.ScrollToBottom()
}

func cleanString(s string) string {
	re := regexp.MustCompile("[^\x20-\x7E\n]+")
	return re.ReplaceAllString(s, "")
}

func showWarningMessageWithConfirmation(g *Gui, warningMessage string, confirmAction binding.DataListener) {
	var wizard *dialogWizard.Wizard

	// 	warningInfoLabel := widget.NewLabel(`By clicking "Proceed," you confirm that you have saved your mnemonic. You will no longer be able to see your mnemonic a second time. Make sure you have securely stored it before proceeding.
	// If you have not please press "Return" and save your mnemonic.`)
	warningInfoLabel := widget.NewLabel(warningMessage)

	warningInfoLabel.Wrapping = fyne.TextWrapWord
	warningInfoLabel.Importance = widget.DangerImportance
	proceedButton := widget.NewButtonWithIcon("Proceed", theme.ConfirmIcon(), func() {
		wizard.Hide()
		confirmAction.DataChanged()
	})
	returnButton := widget.NewButton("Return", func() { wizard.Hide() })

	content := container.NewBorder(
		nil,
		container.NewGridWithColumns(2, proceedButton, returnButton),
		nil,
		nil,
		container.NewVScroll(warningInfoLabel),
	)
	wizard = dialogWizard.NewWizard("WARNING!", content)
	wizard.Show(g.Window)
	wizard.Resize(fyne.NewSize(500, 400))
}

func showMonikerEntryDialog(g *Gui, monikerBinding binding.String, confirmAction binding.DataListener) {
	var wizard *dialogWizard.Wizard

	monikerEntry := widget.NewEntry()
	monikerEntry.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
	monikerEntry.MultiLine = true

	doneButton := widget.NewButton("Claim", func() {
		moniker := monikerEntry.Text
		trimmed := strings.TrimSpace(moniker)
		if trimmed == "" {
			monikerEntry.SetValidationError(fmt.Errorf("moniker cannot be empty"))
			return
		}
		monikerBinding.Set(moniker)
		confirmAction.DataChanged()
		wizard.Hide()
	})
	doneButton.Importance = widget.HighImportance

	cancelButton := widget.NewButton("Cancel", func() { wizard.Hide() })

	content := container.NewBorder(
		nil,
		container.NewVBox(doneButton, cancelButton),
		nil, nil,
		container.NewVScroll(monikerEntry),
	)

	wizard = dialogWizard.NewWizard("Enter your Moniker", content)
	wizard.Show(g.Window)
	wizard.Resize(fyne.NewSize(500, 400))
}
