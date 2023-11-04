package client

import (
	"context"
	"fmt"

	"github.com/ArtemShalinFe/gophkeeper/internal/config"
	"github.com/ArtemShalinFe/gophkeeper/internal/models"
	"github.com/ArtemShalinFe/gophkeeper/internal/server"
	"github.com/ArtemShalinFe/gophkeeper/internal/storage/mem"

	"github.com/rivo/tview"
	"go.uber.org/zap"
)

const (
	// DefaultFieldWidth - The width of the default fields that will be displayed in the interface.
	defaultFieldWidth = 40
)

// TUI - An object that contains everything necessary for the text user interface to work correctly.
type TUI struct {
	app      *tview.Application
	pages    *tview.Pages
	gkclient *server.GKClient
	authUser *models.User
	cache    *mem.MemStorage
	recLimit int
}

// Start - starts graphical text user interface.
func (ui *TUI) Start(ctx context.Context) error {
	ui.app = tview.NewApplication()
	ui.pages = tview.NewPages()
	ui.recLimit = models.DefaultLimit

	log := zap.L()
	cfg := config.NewClientCfg()
	if err := config.ReadEnvClientCfg(cfg); err != nil {
		return fmt.Errorf("an error occured while read config, err: %w", err)
	}
	gkclient, err := server.NewGKClient(ctx, cfg, log)
	if err != nil {
		return fmt.Errorf("an error occure while init gk client, err: %w", err)
	}

	ui.gkclient = gkclient
	ui.cache = mem.NewMemStorage()
	ui.displayUserLoginPage(ctx)

	appStopCh := make(chan error)

	flex := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(ui.pages, 0, 1, true), 0, 1, true)

	go func() {
		appStopCh <- ui.app.SetRoot(flex, true).EnableMouse(true).SetFocus(flex).Run()
	}()

	select {
	case <-ctx.Done():
		ui.app.Stop()
		return <-appStopCh
	case err := <-appStopCh:
		return err
	}
}

// Sync - Synchronizes the client storage cache and the server storage using the version vector mechanism.
func (ui *TUI) Sync(ctx context.Context) {
	if ui.authUser == nil {
		return
	}
	if err := ui.authUser.SyncRecords(ctx, ui.cache, ui.gkclient, defaulTickSync); err != nil {
		ui.displayErr(err.Error())
	}
}
