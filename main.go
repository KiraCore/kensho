package main

import (
	"flag"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/KiraCore/kensho/gui"
	"github.com/KiraCore/kensho/types"
	"github.com/KiraCore/kensho/utils"
)

func main() {
	devMode := flag.Bool("dev", false, "Enable developer mode")
	flag.Parse()

	a := app.NewWithID(types.APP_NAME)
	w := a.NewWindow(types.APP_NAME)
	p := a.Preferences()
	w.SetMaster()
	w.Resize(fyne.NewSize(1024, 768))
	homeFolder, err := utils.InitHomeFolder(types.APP_NAME)
	if err != nil {
		log.Println(err)
	}

	g := gui.Gui{
		DeveloperMode: *devMode,
		Window:        w,
		Version:       a.Metadata().Version,
		Preferences:   p,
		HomeFolder:    homeFolder,
	}
	g.WaitDialog = gui.NewWaitDialog(&g)
	content := g.MakeGui()
	g.Window.SetContent(content)
	a.Lifecycle().SetOnStarted(func() {
		g.ShowConnect()
	})
	w.ShowAndRun()
}
