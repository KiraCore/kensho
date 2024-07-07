package gui

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	dialogWizard "github.com/KiraCore/kensho/gui/dialogs"
	"github.com/KiraCore/kensho/helper/httph"
	"github.com/KiraCore/kensho/types"
)

func showSekaiExecuteDialog(g *Gui) {
	var wizard *dialogWizard.Wizard

	sekaidCmdName := "/sekaid"
	cmdData := binding.NewString()
	cmdEntry := widget.NewEntryWithData(cmdData)
	cmdEntry.MultiLine = true
	cmdEntry.Wrapping = fyne.TextWrapWord

	doneAction := binding.NewDataListener(func() {
		g.WaitDialog.ShowWaitDialog()
		log.Printf("Trying to execute: %v", cmdEntry.Text)
		cmd, _ := cmdData.Get()
		cmd = strings.ReplaceAll(cmd, "\n", " ")
		cmd = fmt.Sprintf("%v %v", sekaidCmdName, cmd)

		cmdArgs := strings.Split(cmd, " ")
		cmdArgs = RemoveEmptyAndWhitespaceStrings(cmdArgs)

		cmdStruct := types.ExecSekaiCommands{Command: "sekaid", ExecArgs: types.ExecArgs{Exec: cmdArgs}}

		payload, err := json.Marshal(cmdStruct)
		if err != nil {
			log.Printf("error when marshaling cmdStruct: %v", err.Error())
			g.WaitDialog.HideWaitDialog()
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
			return
		}
		ctx := context.Background()
		o, err := httph.ExecHttpRequestBySSHTunnel(ctx, g.sshClient, types.SEKIN_EXECUTE_ENDPOINT, "POST", payload)
		if err != nil {
			log.Printf("error when executing cmdStruct: %v", err.Error())
			g.WaitDialog.HideWaitDialog()
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
			return
		}

		log.Printf("output of <%v>:\n%v", cmd, string(o))
		g.WaitDialog.HideWaitDialog()

		showInfoDialog(g, "Out", string(o))
	})

	submitButton := widget.NewButton("Submit", func() {
		log.Printf("Submitting sekai cmd: %v", cmdEntry.Text)
		warningMessage := fmt.Sprintf("Are you sure you want to execute this?\n\nCommand: <%v>\n\nYou cannot revert changes", cmdEntry.Text)
		showWarningMessageWithConfirmation(g, warningMessage, doneAction)

	})
	closeButton := widget.NewButton("Cancel", func() {
		wizard.Hide()
	})
	submitButton.Importance = widget.HighImportance

	hintCmd := "'tx bank send <from_key_or_address> <to_address> 1000ukex --chain-id=chaosnet2 --home=/sekai --fees=100ukex \n--keyring-backend=test --yes --broadcast-mode=block --log_format=json --output=json'  "
	hintText := &widget.TextSegment{Text: hintCmd, Style: widget.RichTextStyleBlockquote}
	welcomeText := &widget.TextSegment{Text: "Example:", Style: widget.RichTextStyle{TextStyle: fyne.TextStyle{Bold: true}}}

	hintTextWidget := widget.NewRichText(welcomeText, hintText)
	// hintTextWidget.Wrapping = fyne.TextWrapWord  // dont use word wrapping, richtext glitches with this one

	entryItem := widget.NewFormItem(sekaidCmdName[1:]+":", container.NewVBox(cmdEntry))

	content := container.NewBorder(
		nil,
		container.NewVBox(submitButton, closeButton),
		nil,
		nil,
		container.NewVBox(widget.NewForm(entryItem), hintTextWidget),
	)
	wizard = dialogWizard.NewWizard("Sekai executor", content)
	wizard.Show(g.Window)
	wizard.Resize(fyne.NewSize(800, 200))

}

func RemoveEmptyAndWhitespaceStrings(input []string) []string {
	var result []string
	for _, str := range input {
		trimmedStr := strings.TrimSpace(str)
		if trimmedStr != "" {
			result = append(result, trimmedStr)
		}
	}
	return result
}
