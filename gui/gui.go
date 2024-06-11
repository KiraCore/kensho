package gui

import (
	"context"
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/crypto/ssh"
)

const appName = "Kensho 1.0"

type Gui struct {
	sshClient               *ssh.Client
	Window                  fyne.Window
	WaitDialog              *WaitDialog
	HomeFolder              string
	Host                    *Host
	ConnectionStatusBinding binding.Bool
	ConnectionCount         int
	Terminal                Terminal
	LogCtx                  context.Context
	LogCtxCancel            context.CancelFunc
	NodeInfo                nodeInfoScreen
	TxExec                  TxExecBinding

	Version string
}

type TxExecBinding struct {
	TxExecutionStatusBinding binding.Bool
	TxDoneListener           binding.DataListener
}
type Host struct {
	IP string
}

func (g *Gui) MakeGui() fyne.CanvasObject {
	title := widget.NewLabel(appName)
	info := widget.NewLabel("Welcome to  Kensho. Navigate trough panel on the left side")
	mainWindow := container.NewStack()

	reconnectButton := widget.NewButton("Reconnect", func() {
		g.ShowConnect()
	})
	reconnectButton.Hide()
	reconnectButton.Importance = widget.DangerImportance

	loadingData := binding.NewFloat()

	loadWidget := widget.NewProgressBarWithData(loadingData)
	g.TxExec.TxDoneListener = binding.NewDataListener(func() {})
	txExecLoadingWidget := container.NewStack(loadWidget)
	txExecLoadingWidget.Hide()
	g.TxExec.TxExecutionStatusBinding = binding.NewBool()
	g.TxExec.TxExecutionStatusBinding.AddListener(binding.NewDataListener(func() {
		state, _ := g.TxExec.TxExecutionStatusBinding.Get()
		log.Println("TxExecutionStatusBinding state:", state)
		if state {
			txExecLoadingWidget.Show()
			var maxRange float64 = 40
			for i := range int(maxRange) {
				time.Sleep(time.Second * 1)
				var percentage float64 = ((float64(i) * 100) / maxRange) * 0.01
				// log.Println("wait tx state:", percentage)
				loadingData.Set(float64(percentage))
				// txExecLoadingWidget.Refresh()
				loadWidget.SetValue(float64(percentage))
				loadWidget.Refresh()
			}
			g.TxExec.TxExecutionStatusBinding.Set(false)
			g.TxExec.TxDoneListener.DataChanged()
			txExecLoadingWidget.Hide()
		}
	}))

	tab := container.NewBorder(container.NewVBox(title, info), container.NewVBox(txExecLoadingWidget, reconnectButton), nil, nil, mainWindow)

	g.ConnectionStatusBinding = binding.NewBool()
	g.ConnectionStatusBinding.AddListener(binding.NewDataListener(func() {
		state, _ := g.ConnectionStatusBinding.Get()
		if state {
			g.Window.SetTitle(fmt.Sprintf("%v (connected) - %v", appName, g.Host.IP))
			reconnectButton.Hide()
			tab.Refresh()
		} else {
			g.Window.SetTitle(fmt.Sprintf("%v (not connected)", appName))
			if g.ConnectionCount != 0 {
				log.Println("reconnect button show triggered, connection count:", g.ConnectionCount)
				reconnectButton.Show()
			}
		}
	}))

	setTab := func(t Tab) {
		title.SetText(t.Title)
		info.SetText(t.Info)
		mainWindow.Objects = []fyne.CanvasObject{t.View(g.Window, g)}
	}
	versionData := binding.NewString()
	versionData.Set(fmt.Sprintf("Version: %v", g.Version))
	versionLabel := widget.NewLabelWithData(versionData)
	versionLabel.Alignment = fyne.TextAlignCenter
	menuAndTab := container.NewHSplit(container.NewBorder(nil, versionLabel, nil, nil, g.makeNav(setTab)), tab)
	menuAndTab.Offset = 0.2
	return menuAndTab

}

func (g *Gui) makeNav(setTab func(t Tab)) fyne.CanvasObject {
	a := fyne.CurrentApp()
	const preferenceCurrent = "nav"

	tree := &widget.Tree{
		ChildUIDs: func(uid string) []string {
			return TabsIndex[uid]
		},
		IsBranch: func(uid string) bool {
			children, ok := TabsIndex[uid]

			return ok && len(children) > 0
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Collection Widgets")
		},
		UpdateNode: func(uid string, branch bool, obj fyne.CanvasObject) {
			t, ok := Tabs[uid]
			if !ok {
				fyne.LogError("Missing panel: "+uid, nil)
				return
			}
			obj.(*widget.Label).SetText(t.Title)
			// if unsupportedTutorial(t) {
			// 	obj.(*widget.Label).TextStyle = fyne.TextStyle{Italic: true}
			// } else {
			// 	obj.(*widget.Label).TextStyle = fyne.TextStyle{}
			// }
			obj.(*widget.Label).TextStyle = fyne.TextStyle{}
		},
		OnSelected: func(uid string) {
			if t, ok := Tabs[uid]; ok {
				// if unsupportedTutorial(t) {
				// 	return
				// }
				// fmt.Println(uid)
				a.Preferences().SetString(preferenceCurrent, uid)
				setTab(t)
			}
		},
		OnUnselected: func(uid widget.TreeNodeID) {
			switch uid {
			case "logs":
				log.Println("Unselected: ", uid)
				g.LogCtxCancel()
			case "nodeInfo":
				log.Println("Unselected: ", uid)
				g.NodeInfo.ctxCancel()
			}

		},
	}

	return tree
}
