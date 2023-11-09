package client

import (
	"time"

	"github.com/rivo/tview"
)

const (
	buttonQuinDesc     = "Quit"
	buttonCancelDesc   = "Cancel"
	buttonRegisterDesc = "Register"
	buttonLoginDesc    = "Login"
	buttonOkDesc       = "Ok"
	buttonUpdate       = "Update"
)

func (ui *TUI) displayQuitModal() {
	name := "Exit question"

	modal := tview.NewModal().
		SetText("Do you want to quit the application?").
		AddButtons([]string{buttonCancelDesc, buttonQuinDesc}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == buttonQuinDesc {
				ui.app.Stop()
			}
			if buttonLabel == buttonCancelDesc {
				ui.pages.RemovePage(name)
			}
		})

	ui.pages.AddPage(name, modal, true, true)
}

func (ui *TUI) displayErr(text string) {
	name := "Error"

	modal := tview.NewModal().
		SetText(text).
		AddButtons([]string{buttonOkDesc}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == buttonOkDesc {
				ui.pages.RemovePage(name)
			}
		})

	ui.pages.AddPage(name, modal, true, true)
}

// statusSetup - sets sync status text.
//
// If interval is zero or less sets text status permanently.
func (ui *TUI) statusSetup(text string, interval int) {
	ui.syncStatus.SetText(text)

	if interval > 0 {
		ui.statusCleanup(time.Duration(interval) * time.Second)
	}
}

// statusCleanup - clean text in sync status string.
func (ui *TUI) statusCleanup(interval time.Duration) {
	timer := time.NewTimer(interval)

	go func() {
		<-timer.C
		ui.syncStatus.Clear()
	}()
}
