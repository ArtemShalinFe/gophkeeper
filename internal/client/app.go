package client

import (
	"context"
	"log"

	"go.uber.org/zap"
)

// App - The object that is used to launch the TUI interface.
type App struct {
	// ui - text user interface.
	ui *TUI
	// log - app logger.
	log *zap.Logger
}

// NewApp - Object Constructor.
func NewApp(log *zap.Logger) *App {
	return &App{
		ui:  &TUI{},
		log: log,
	}
}

// Start - The program execution starts with this function.
func (a *App) Start(ctx context.Context) {
	if err := a.ui.Start(ctx); err != nil {
		log.Printf("an unexpected error occured, err: %v", err)
	}
}
