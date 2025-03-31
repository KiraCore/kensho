package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/KiraCore/kensho/helper/httph"
)

func makeCfgEditorScreen(_ fyne.Window, g *Gui) fyne.CanvasObject {
	appTomlTab := container.NewTabItem("app.toml", makeTextEditTab(
		g,
		func(cfg string) error {
			err := httph.SetAppTomlConfig(g.sshClient, cfg, 8282)
			if err != nil {
				return err
			}
			return nil
		},
		func() (string, error) {
			cfg, err := httph.GetAppTomlConfig(g.sshClient, 8282)
			if err != nil {
				return "", err
			}
			return cfg, nil
		},
	))
	configTomlTab := container.NewTabItem("config.toml", makeTextEditTab(
		g,
		func(cfg string) error {
			err := httph.SetConfigTomlConfig(g.sshClient, cfg, 8282)
			if err != nil {
				return err
			}
			return nil
		},
		func() (string, error) {
			cfg, err := httph.GetConfigTomlConfig(g.sshClient, 8282)
			if err != nil {
				return "", err
			}
			return cfg, nil
		},
	))
	tabsMenu := container.NewAppTabs(appTomlTab, configTomlTab)

	return tabsMenu
}

func makeTextEditTab(g *Gui, saveFile func(string) error, getFile func() (string, error)) fyne.CanvasObject {
	configBinding := binding.NewString()
	configEditor := widget.NewEntryWithData(configBinding)
	configEditor.MultiLine = true
	configEditor.Disable()

	const editButtonEnabledState = "Disable editing"
	const editButtonDisabledState = "Enable editing"

	editButton := widget.NewButton(editButtonDisabledState, func() {})
	saveButton := widget.NewButton("Save", func() {})
	saveButton.Disable()
	refreshButton := widget.NewButton("Refresh", func() {})

	refreshFunc := func() {
		g.WaitDialog.ShowWaitDialog()
		cfg, err := getFile()
		if err != nil {
			g.WaitDialog.HideWaitDialog()
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
		}
		configBinding.Set(cfg)

		g.WaitDialog.HideWaitDialog()
	}

	saveFunc := func() {
		g.WaitDialog.ShowWaitDialog()
		cfg, _ := configBinding.Get()
		err := saveFile(cfg)
		if err != nil {
			g.WaitDialog.HideWaitDialog()
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
		}

		g.WaitDialog.HideWaitDialog()
		refreshFunc()
	}

	editFunc := func() {
		if configEditor.Disabled() {
			editButton.SetText(editButtonEnabledState)
			saveButton.Enable()
			configEditor.Enable()
		} else {
			saveButton.Disable()
			editButton.SetText(editButtonDisabledState)
			configEditor.Disable()
		}
	}

	editButton.OnTapped = editFunc
	refreshButton.OnTapped = refreshFunc
	saveButton.OnTapped = saveFunc

	buttonsContainer := container.NewVBox(
		container.NewGridWithColumns(2, editButton, saveButton),
		refreshButton,
	)
	refreshFunc()
	return container.NewBorder(nil, buttonsContainer, nil, nil, configEditor)
}
