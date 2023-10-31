package client

import (
	"context"
	"time"

	"github.com/ArtemShalinFe/gophkeeper/internal/models"
	"github.com/rivo/tview"
)

const defaulTickSync = 5

func (ui *TUI) displayUserLoginPage(ctx context.Context) {
	var userDTO models.UserDTO
	form := tview.NewForm().
		AddInputField(fnUsername, "", defaultFieldWidth, nil, func(v string) {
			userDTO.Login = v
		}).
		AddPasswordField(fnPassword, "", defaultFieldWidth, '*', func(v string) {
			userDTO.Password = v
		}).
		AddButton(buttonLoginDesc, func() {
			u, err := userDTO.GetUser(ctx, ui.gkclient)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}
			ui.authUser = u

			if err := ui.cache.AddUserRecordStorage(u.ID); err != nil {
				ui.displayErr(err.Error())
			}

			ctxT, cancel := context.WithTimeout(ctx, time.Second*1)
			defer cancel()

			ui.Sync(ctxT)

			ui.displayRecords(ctx, 0, ui.recLimit)

			go ui.Sync(ctx)
		}).
		AddButton(buttonRegisterDesc, func() {
			u, err := userDTO.AddUser(ctx, ui.gkclient)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}
			ui.authUser = u

			if err := ui.cache.AddUserRecordStorage(u.ID); err != nil {
				ui.displayErr(err.Error())
			}

			ctxT, cancel := context.WithTimeout(ctx, time.Second*1)
			defer cancel()

			ui.Sync(ctxT)

			ui.displayRecords(ctx, 0, ui.recLimit)

			go ui.Sync(ctx)
		}).
		AddButton(buttonQuinDesc, ui.displayQuitModal)

	form.SetBorder(true).SetTitle(" GophKeeper ").
		SetTitleAlign(tview.AlignLeft)

	ui.pages.AddPage(loginPage, form, true, true)
}
