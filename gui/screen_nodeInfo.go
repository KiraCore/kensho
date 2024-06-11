package gui

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/atotto/clipboard"

	"github.com/KiraCore/kensho/helper/httph"
	"github.com/KiraCore/kensho/types"
	"github.com/KiraCore/kensho/types/endpoint/shidai"
)

type nodeInfoScreen struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
}

func makeNodeInfoScreen(_ fyne.Window, g *Gui) fyne.CanvasObject {
	return makeNodeInfoTab(g)
}

func makeNodeInfoTab(g *Gui) fyne.CanvasObject {

	ctx := context.Background()
	g.NodeInfo.ctx, g.NodeInfo.ctxCancel = context.WithCancel(ctx)

	var claimSeat bool
	var refreshBinding binding.DataListener

	validatorControlButton := widget.NewButton("", func() {})
	validatorControlButton.Disable()
	validatorControlButton.Hide()

	latestBlockData := binding.NewString()
	latestBlockLabel := widget.NewLabelWithData(latestBlockData)

	// validator address box
	validatorAddressData := binding.NewString()
	validatorAddressLabel := widget.NewLabelWithData(validatorAddressData)
	validatorAddressCopyButton := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		data, err := validatorAddressData.Get()
		if err != nil {
			log.Println(err)
			return
		}

		err = clipboard.WriteAll(data)
		if err != nil {
			return
		}
	})

	//validator status (active, paused, etc...)
	validatorStatusData := binding.NewString()
	validatorStatusLabel := widget.NewLabelWithData(validatorStatusData)

	// nodeID
	nodeIDData := binding.NewString()
	nodeIDLabel := widget.NewLabelWithData(nodeIDData)

	topData := binding.NewString()
	topLabel := widget.NewLabelWithData(topData)

	// miss chance box
	missChanceData := binding.NewString()
	missChanceLabel := widget.NewLabelWithData(missChanceData)

	lastProducedBlockData := binding.NewString()
	lastProducedLabel := widget.NewLabelWithData(lastProducedBlockData)

	// miss Chance Confidence box
	missChanceConfidenceData := binding.NewString()
	missChanceConfidenceLabel := widget.NewLabelWithData(missChanceConfidenceData)

	startHeightData := binding.NewString()
	startHeightLabel := widget.NewLabelWithData(startHeightData)

	producedBlocksData := binding.NewString()
	producedBlocksLabel := widget.NewLabelWithData(producedBlocksData)

	monikerData := binding.NewString()
	monikerLabel := widget.NewLabelWithData(monikerData)

	genesisChecksumData := binding.NewString()
	genesisChecksumLabel := widget.NewLabelWithData(genesisChecksumData)

	nodeCatchingData := binding.NewString()
	nodeCatchingLabel := widget.NewLabelWithData(nodeCatchingData)

	// Numerical validators status
	activeValidatorsData := binding.NewInt()
	activeValidatorsLabel := widget.NewLabelWithData(binding.IntToString(activeValidatorsData))
	activeValidatorsLabel.Alignment = fyne.TextAlignCenter
	activeValidatorsInfoText := widget.NewLabel("Active:")
	activeValidatorsInfoText.TextStyle.Bold = true
	activeValidatorsBox := container.NewVBox(
		activeValidatorsInfoText, activeValidatorsLabel,
	)

	pausedValidatorsData := binding.NewInt()
	pausedValidatorsLabel := widget.NewLabelWithData(binding.IntToString(pausedValidatorsData))
	pausedValidatorsLabel.Alignment = fyne.TextAlignCenter
	pausedValidatorsInfoText := widget.NewLabel("Paused:")
	pausedValidatorsInfoText.TextStyle.Bold = true
	pausedValidatorsBox := container.NewVBox(
		pausedValidatorsInfoText, pausedValidatorsLabel,
	)

	inactiveValidatorsData := binding.NewInt()
	inactiveValidatorsLabel := widget.NewLabelWithData(binding.IntToString(inactiveValidatorsData))
	inactiveValidatorsLabel.Alignment = fyne.TextAlignCenter
	inactiveValidatorsInfoText := widget.NewLabel("Inactive:")
	inactiveValidatorsInfoText.TextStyle.Bold = true

	inactiveValidatorsBox := container.NewVBox(
		inactiveValidatorsInfoText, inactiveValidatorsLabel,
	)

	jailedValidatorsData := binding.NewInt()
	jailedValidatorsLabel := widget.NewLabelWithData(binding.IntToString(jailedValidatorsData))
	jailedValidatorsLabel.Alignment = fyne.TextAlignCenter

	jailedValidatorsInfoText := widget.NewLabel("Jailed:")
	jailedValidatorsInfoText.TextStyle.Bold = true

	jailedValidatorsBox := container.NewVBox(
		jailedValidatorsInfoText, jailedValidatorsLabel,
	)

	waitingValidatorsData := binding.NewInt()
	waitingValidatorsLabel := widget.NewLabelWithData(binding.IntToString(waitingValidatorsData))
	waitingValidatorsLabel.Alignment = fyne.TextAlignCenter

	waitingValidatorsInfoText := widget.NewLabel("Waiting:")
	waitingValidatorsInfoText.TextStyle.Bold = true

	waitingValidatorsBox := container.NewVBox(
		waitingValidatorsInfoText, waitingValidatorsLabel,
	)

	chainIDData := binding.NewString()
	chainIDLabel := widget.NewLabelWithData(chainIDData)

	streakData := binding.NewString()
	streakLabel := widget.NewLabelWithData(streakData)

	valuesForm := widget.NewForm(
		widget.NewFormItem("ChainID:", chainIDLabel),
		widget.NewFormItem("Moniker:", monikerLabel),
		widget.NewFormItem("Val.Status:", validatorStatusLabel),
		widget.NewFormItem("Node Status:", nodeCatchingLabel),
		widget.NewFormItem("Block:", latestBlockLabel),
		widget.NewFormItem("Latest Block:", lastProducedLabel),
		widget.NewFormItem("Produced:", producedBlocksLabel),
		widget.NewFormItem("Streak:", streakLabel),
		widget.NewFormItem("Rank:", topLabel),
		widget.NewFormItem("Miss:", missChanceLabel),
		widget.NewFormItem("Miss conf.", missChanceConfidenceLabel),
		widget.NewFormItem("Start Height:", startHeightLabel),
		widget.NewFormItem("Val.Addr:", container.NewHBox(validatorAddressLabel, validatorAddressCopyButton)),
		widget.NewFormItem("Val.NodeID:", nodeIDLabel),
		widget.NewFormItem("Gen.SHA256:", genesisChecksumLabel),
	)
	execFunc := func(args types.ExecSekaiCmd) {
		g.TxExec.TxExecutionStatusBinding.Set(true)

		request := types.RequestTXPayload{Command: "tx", Args: args}
		payload, err := json.Marshal(request)

		log.Printf("Executing: %+v", args)
		if err != nil {
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
		}
		out, err := httph.ExecHttpRequestBySSHTunnel(g.sshClient, types.SEKIN_EXECUTE_ENDPOINT, "POST", payload)
		if err != nil {
			log.Println("ERROR when executing payload:", err.Error())
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
			return
		}

		log.Println("payload execution out:", string(out))
		refreshBinding.DataChanged()
	}

	monikerEntryData := binding.NewString()
	claimDataListener := binding.NewDataListener(func() {
		moniker, _ := monikerEntryData.Get()
		execFunc(types.ExecSekaiCmd{TX: types.ClaimValidatorSeat, Moniker: moniker})
	})

	claimValidatorSeatFunc := func() {
		//claim seat
		showMonikerEntryDialog(g, monikerEntryData, claimDataListener)
	}
	pauseValidatorFunc := func() {
		// pause
		execFunc(types.ExecSekaiCmd{TX: types.Pause})
	}
	unpauseValidatorFunc := func() {
		// unpause tx
		execFunc(types.ExecSekaiCmd{TX: types.Unpause})
	}
	activateValidatorFunc := func() {
		// activate
		execFunc(types.ExecSekaiCmd{TX: types.Activate})
	}

	errBinding := binding.NewUntyped()

	refreshScreen := func() {
		g.WaitDialog.ShowWaitDialog()
		defer g.WaitDialog.HideWaitDialog()
		dashboardData, err := httph.GetDashboardInfo(g.sshClient, 8282)
		if err != nil {
			errBinding.Set(fmt.Errorf("ERROR: getting dashboard info: %w", err))
			return
		}

		log.Printf("dashboard %+v", dashboardData)

		txExec, _ := g.TxExec.TxExecutionStatusBinding.Get()

		if dashboardData.ValidatorStatus != "Unknown" || dashboardData.SeatClaimAvailable {
			validatorControlButton.Enable()
			log.Println("txExec =", txExec)
			if txExec {
				if validatorControlButton.Hidden {
					validatorControlButton.Show()
				}
				validatorControlButton.Disable()
			} else {
				if validatorControlButton.Hidden {
					validatorControlButton.Show()
				}
				if validatorControlButton.Disabled() {
					validatorControlButton.Enable()
				}

				status := strings.ToUpper(dashboardData.ValidatorStatus)
				claimSeat = dashboardData.SeatClaimAvailable
				if claimSeat {
					if validatorControlButton.Disabled() {
						validatorControlButton.Enable()
					}
					if validatorControlButton.Hidden {
						validatorControlButton.Show()
					}
					validatorControlButton.SetText("Claim Validator Seat")
					validatorControlButton.OnTapped = claimValidatorSeatFunc
				} else {
					switch status {
					case string(shidai.Active):
						if validatorControlButton.Disabled() {
							validatorControlButton.Enable()
						}
						if validatorControlButton.Hidden {
							validatorControlButton.Show()
						}
						validatorControlButton.SetText("Pause")
						validatorControlButton.OnTapped = pauseValidatorFunc
					case string(shidai.Paused):
						if validatorControlButton.Disabled() {
							validatorControlButton.Enable()
						}
						if validatorControlButton.Hidden {
							validatorControlButton.Show()
						}
						validatorControlButton.SetText("Unpause")
						validatorControlButton.OnTapped = unpauseValidatorFunc

					case string(shidai.Inactive):
						if validatorControlButton.Disabled() {
							validatorControlButton.Enable()
						}
						if validatorControlButton.Hidden {
							validatorControlButton.Show()
						}
						validatorControlButton.SetText("Activate")
						validatorControlButton.OnTapped = activateValidatorFunc
					}
				}

			}
		}
		validatorControlButton.Refresh()
		nodeIDData.Set(dashboardData.NodeID)
		topData.Set(dashboardData.Top)
		validatorAddressData.Set(dashboardData.ValidatorAddress)
		missChanceData.Set(dashboardData.Mischance)

		validatorStatusData.Set(dashboardData.ValidatorStatus)

		latestBlockData.Set(dashboardData.Blocks)
		lastProducedBlockData.Set(dashboardData.LastProducedBlock)
		missChanceConfidenceData.Set(dashboardData.MischanceConfidence)
		producedBlocksData.Set(dashboardData.ProducedBlocks)

		startHeightData.Set(dashboardData.StartHeight)
		monikerData.Set(dashboardData.Moniker)
		activeValidatorsData.Set(dashboardData.ActiveValidators)
		pausedValidatorsData.Set(dashboardData.PausedValidators)
		inactiveValidatorsData.Set(dashboardData.InactiveValidators)
		jailedValidatorsData.Set(dashboardData.JailedValidators)
		waitingValidatorsData.Set(dashboardData.WaitingValidators)
		genesisChecksumData.Set(dashboardData.GenesisChecksum)
		chainIDData.Set(dashboardData.ChainID)
		streakData.Set(dashboardData.Streak)

		if dashboardData.CatchingUp {
			nodeCatchingData.Set("Syncing")
		} else {
			nodeCatchingData.Set("Running")
		}
	}
	refreshBinding = binding.NewDataListener(func() {
		refreshScreen()
		err, _ := errBinding.Get()
		if err != nil {
			if e, ok := err.(error); ok {
				// g.showErrorDialog(e, binding.NewDataListener(func() {}))
				log.Printf("Refreshing unsuccessful, reason: %v", e.Error())
			}
		}
	})
	g.TxExec.TxDoneListener = binding.NewDataListener(func() { refreshBinding.DataChanged() })

	go func() {
		refreshTime := 20 * time.Second
		log.Printf("Starting goroutine with refresh rate %v", refreshTime)
		timer := time.NewTimer(refreshTime)
		defer timer.Stop() // Clean up the timer when the goroutine ends

		refreshBinding.DataChanged()
		for {
			select {
			case <-g.NodeInfo.ctx.Done():
				log.Printf("Ending nodeInfo refresh goroutine")
				return
			case <-timer.C:
				refreshBinding.DataChanged()
				timer.Reset(refreshTime) // Reset the timer
			}
		}
	}()

	// refreshButton := widget.NewButton("Refresh", refreshBinding.DataChanged)
	// sendSekaiCommandButton := widget.NewButton("Execute sekai command", func() { showSekaiExecuteDialog(g) })
	mainInfo := container.NewVScroll(
		container.NewVBox(
			widget.NewSeparator(),
			valuesForm,
		),
	)

	validatorsTopPart := container.NewHBox(activeValidatorsBox, pausedValidatorsBox, inactiveValidatorsBox, jailedValidatorsBox, waitingValidatorsBox)

	// return container.NewBorder(nil, refreshButton, nil, validatorsRightPart, mainInfo)
	return container.NewBorder(container.NewCenter(validatorsTopPart), validatorControlButton, nil, nil, mainInfo)
}
