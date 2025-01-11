package main

import (
	"flag"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/KiraCore/kensho/gui"
)

func main() {
	devMode := flag.Bool("dev", false, "Enable developer mode")
	flag.Parse()

	a := app.NewWithID("Kensho")
	w := a.NewWindow("Kensho")
	p := a.Preferences()
	w.SetMaster()
	w.Resize(fyne.NewSize(1024, 768))
	g := gui.Gui{
		DeveloperMode: *devMode,
		Window:        w,
		Version:       a.Metadata().Version,
		Preferences:   p,
	}
	g.WaitDialog = gui.NewWaitDialog(&g)
	content := g.MakeGui()
	g.Window.SetContent(content)
	a.Lifecycle().SetOnStarted(func() {
		g.ShowConnect()
	})
	w.ShowAndRun()
}
