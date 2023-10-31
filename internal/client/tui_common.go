package client

import (
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
