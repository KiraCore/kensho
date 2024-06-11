package gui

import (
	"encoding/json"
	"log"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/KiraCore/kensho/helper/httph"
	"github.com/KiraCore/kensho/types"
)

func makeStatusScreen(_ fyne.Window, g *Gui) fyne.CanvasObject {
	const STATUS_Unavailable = "Unavailable"
	const STATUS_Running = "Running"

	interxStatusCheck := binding.NewBool()
	interxInfraCheck := binding.NewBool()
	interxStatusInfo := widget.NewLabel("")
	interxInfoBox := container.NewHBox(
		widget.NewLabel("Interx:"),
		interxStatusInfo,
	)

	shidaiStatusCheck := binding.NewBool()
	shidaiInfraCheck := binding.NewBool()
	shidaiStatusInfo := widget.NewLabel("")
	shidaiInfoBox := container.NewHBox(
		widget.NewLabel("Shidai:"),
		shidaiStatusInfo,
	)

	sekaiStatusCheck := binding.NewBool()
	sekaiInfraCheck := binding.NewBool()
	sekaiStatusInfo := widget.NewLabel("")
	sekaiInfoBox := container.NewHBox(
		widget.NewLabel("Sekai:"),
		sekaiStatusInfo,
	)

	var dataListenerForSuccesses binding.DataListener
	deployButton := widget.NewButton("Deploy", func() {
		showDeployDialog(g, dataListenerForSuccesses, shidaiInfraCheck)
	})
	deployButton.Disable()

	checkInterxStatus := func() {
		_, err := httph.GetInterxStatus(g.Host.IP, strconv.Itoa(types.DEFAULT_INTERX_PORT))
		if err != nil {
			log.Printf("ERROR getting interx status: %v", err)
			interxStatusInfo.SetText(STATUS_Unavailable)
			interxStatusCheck.Set(false)

		} else {
			interxStatusCheck.Set(true)
			interxStatusInfo.SetText(STATUS_Running)
		}

	}

	checkShidaiStatus := func() {
		shidaiStatus, err := httph.GetShidaiStatus(g.sshClient, types.DEFAULT_SHIDAI_PORT)
		if err != nil {
			log.Printf("ERROR: %v", err)
			shidaiStatusInfo.SetText(STATUS_Unavailable)
			shidaiStatusCheck.Set(false)
			shidaiInfraCheck.Set(false)

		} else {
			log.Println("switching  shidai state")
			shidaiStatusInfo.SetText(STATUS_Running)
			shidaiInfraCheck.Set(true)
			sekaiInfraCheck.Set(shidaiStatus.Sekai.Infra)
			interxInfraCheck.Set(shidaiStatus.Interx.Infra)
			shidaiStatusCheck.Set(true)

		}
	}

	checkSekaiStatus := func() {
		_, err := httph.GetSekaiStatus(g.Host.IP, "26657")
		if err != nil {
			log.Printf("ERROR: %v", err)
			sekaiStatusInfo.SetText(STATUS_Unavailable)
			sekaiStatusCheck.Set(false)

		} else {
			sekaiStatusInfo.SetText(STATUS_Running)
			sekaiStatusCheck.Set(true)
		}
	}
	startButton := widget.NewButton("Start", func() {})
	refresh := func() {
		g.WaitDialog.ShowWaitDialog()
		checkInterxStatus()
		checkShidaiStatus()
		checkSekaiStatus()
		shidaiCheck, _ := shidaiStatusCheck.Get()
		sekaiCheck, _ := sekaiStatusCheck.Get()
		interxCheck, _ := interxStatusCheck.Get()

		shidaiInfra, _ := shidaiInfraCheck.Get()
		sekaiInfra, _ := sekaiInfraCheck.Get()
		interxInfra, _ := interxInfraCheck.Get()
		log.Printf("CHECKS: shidaiCheck:%v sekaiCheck:%v interxCheck:%v shidaiInfra:%v sekaiInfra:%v interxInfra:%v",
			shidaiCheck, sekaiCheck, interxCheck, shidaiInfra, sekaiInfra, interxInfra)

		// TODO: first maybe we should try to restart first if shidai is not running
		var deployButtonCheck bool

		if !shidaiCheck {
			deployButtonCheck = true
			log.Println("1st deploy check set", deployButtonCheck)
		} else {
			deployButtonCheck = false
		}
		if !deployButtonCheck {
			if shidaiInfra && sekaiInfra && interxInfra && (shidaiCheck && !sekaiCheck && !interxCheck) {
				startButton.Enable()
				log.Println("start button enabled")
			} else {
				startButton.Disable()
				log.Println("start button disabled")
			}
		}

		log.Println("enable state: ", deployButtonCheck)
		if deployButtonCheck {
			deployButton.Enable()
		}

		defer g.WaitDialog.HideWaitDialog()
	}
	startButton.OnTapped = func() {
		g.WaitDialog.ShowWaitDialog()
		var payloadStruct = types.RequestDeployPayload{
			Command: "start",
		}
		payload, err := json.Marshal(payloadStruct)
		if err != nil {
			log.Println("ERROR when executing payload:", err.Error())
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
			return
		}
		out, err := httph.ExecHttpRequestBySSHTunnel(g.sshClient, types.SEKIN_EXECUTE_ENDPOINT, "POST", payload)
		if err != nil {
			log.Println("ERROR when executing payload:", err.Error())
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
			return
		}
		log.Println("START out:", string(out))
		g.WaitDialog.HideWaitDialog()
		refresh()
	}
	startButton.Disable()
	refreshButton := widget.NewButton("Refresh", func() {
		refresh()
	})

	dataListenerForSuccesses = binding.NewDataListener(func() {
		log.Println("triggering dataListenerForSuccesses")

		deployButton.Disable()
		refresh()
	})
	defer refresh()
	return container.NewBorder(nil,
		container.NewVBox(startButton,
			deployButton,
			widget.NewSeparator(),
			refreshButton), nil, nil,
		container.NewVBox(
			widget.NewSeparator(),
			interxInfoBox,
			sekaiInfoBox,
			shidaiInfoBox,
			widget.NewSeparator(),
		))

}
