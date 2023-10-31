package client

import (
	"context"
	"log"

	"go.uber.org/zap"
)

type App struct {
	ui  *TUI
	log *zap.Logger
}

func NewApp(log *zap.Logger) *App {
	return &App{
		ui:  &TUI{},
		log: log,
	}
}

func (a *App) Start(ctx context.Context) {
	if err := a.ui.Start(ctx); err != nil {
		log.Printf("an unexpected error occured, err: %v", err)
	}
}
