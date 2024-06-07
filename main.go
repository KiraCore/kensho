package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/KiraCore/kensho/gui"
)

func main() {
	a := app.NewWithID("Kensho")
	w := a.NewWindow("Kensho")
	w.SetMaster()
	w.Resize(fyne.NewSize(1024, 768))
	g := gui.Gui{
		Window: w,
	}
	g.WaitDialog = gui.NewWaitDialog(&g)
	content := g.MakeGui()
	g.Window.SetContent(content)
	a.Lifecycle().SetOnStarted(func() {
		g.ShowConnect(false)
	})
	w.ShowAndRun()
}
