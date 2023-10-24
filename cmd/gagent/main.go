package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/ArtemShalinFe/gophkeeper/internal/agent"
	"github.com/ArtemShalinFe/gophkeeper/internal/config"
	"github.com/ArtemShalinFe/gophkeeper/internal/server"
	"github.com/ArtemShalinFe/gophkeeper/internal/storage/mem"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("an occured fatal error, err: %v", err)
	}
}

func run() error {
	ctx, cancelCtx := signal.NotifyContext(context.Background(),
		os.Interrupt,
		os.Kill,
		syscall.SIGQUIT)

	defer cancelCtx()

	componentsErrs := make(chan error, 1)
	go func() {
		select {
		case <-ctx.Done():
		case err := <-componentsErrs:
			zap.L().Error("an unexpected error occurred in one of the application components", zap.Error(err))
			cancelCtx()
		}
	}()

	wg := &sync.WaitGroup{}

	log, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("an error occured while init logger err: %w ", err)
	}

	wg.Add(1)
	go func(log *zap.Logger, errs chan<- error) {
		defer wg.Done()
		<-ctx.Done()
		if err := log.Sync(); err != nil {
			if runtime.GOOS == "darwin" {
				errs <- nil
			} else {
				errs <- fmt.Errorf("cannot flush buffered log entries err: %w", err)
			}
		}
	}(log, componentsErrs)

	cfg := config.NewAgentCfg()
	if err := config.ReadEnvAgentCfg(cfg); err != nil {
		return fmt.Errorf("an error occured when reading environment for server config, err: %w", err)
	}

	gkclient, err := server.NewGKClient(ctx, cfg, log)
	if err != nil {
		return fmt.Errorf("an error occured while init gkeeper client, err: %w", err)
	}
	if err := gkclient.SetupConn(ctx); err != nil {
		return fmt.Errorf("an error occured while init conn to server client, err: %w", err)
	}

	gagent := agent.InitAgent(ctx, *cfg, gkclient, gkclient, mem.NewMemStorage(), log)

	wg.Add(1)
	go func(gagent *agent.Agent, errs chan<- error) {
		if err := gagent.ListenAndServe(); err != nil {
			errs <- fmt.Errorf("an error occured while listen and serve, err: %w", err)
		}
	}(gagent, componentsErrs)

	wg.Add(1)
	go func(gagent *agent.Agent, errs chan<- error) {
		defer log.Error("agent has been shutdown")
		defer wg.Done()
		<-ctx.Done()

		const timeoutServerShutdown = time.Second * 10

		shutdownTimeoutCtx, cancelShutdownTimeoutCtx := context.WithTimeout(context.Background(), timeoutServerShutdown)
		defer cancelShutdownTimeoutCtx()
		if err := gagent.Shutdown(shutdownTimeoutCtx); err != nil {
			errs <- fmt.Errorf("an error occurred during agent shutdown: %w", err)
		}
	}(gagent, componentsErrs)

	defer func() {
		wg.Wait()
	}()

	go func() {
		const timeoutShutdown = time.Second * 30
		ctx, cancelCtx := context.WithTimeout(context.Background(), timeoutShutdown)
		defer cancelCtx()

		<-ctx.Done()
		log.Fatal("failed to gracefully shutdown the service")
	}()

	return nil
}
