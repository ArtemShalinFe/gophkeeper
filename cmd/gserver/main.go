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

	"github.com/ArtemShalinFe/gophkeeper/internal/config"
	"github.com/ArtemShalinFe/gophkeeper/internal/server"
	"github.com/ArtemShalinFe/gophkeeper/internal/storage/sql"
)

const timeoutShutdown = time.Second * 30
const timeoutServerShutdown = time.Second * 10
const componentsCount = 3

func main() {
	if err := run(); err != nil {
		log.Fatalf("an occured fatal error, err: %v", err)
	}
}

func run() error {
	ctx, cancelCtx := signal.NotifyContext(context.Background(),
		os.Interrupt,
		os.Kill,
		syscall.SIGQUIT,
	)

	defer cancelCtx()

	log, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("an error occured while init logger err: %w ", err)
	}

	componentsErrs := make(chan error, componentsCount)
	go func(log *zap.Logger) {
		select {
		case <-ctx.Done():
		case err := <-componentsErrs:
			errTml := "an unexpected error occurred in one of the application components"
			log.Error(errTml, zap.Error(err))
			cancelCtx()
			for err := range componentsErrs {
				log.Error(errTml, zap.Error(err))
			}
		}
	}(log)

	wg := &sync.WaitGroup{}

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

	cfg := config.NewServerCfg()
	if err := config.ReadEnvServerCfg(cfg); err != nil {
		return fmt.Errorf("an error occured when reading environment for server config, err: %w", err)
	}

	log.Info("server will be running", zap.String("Addr", cfg.Addr))

	db, err := sql.NewDB(ctx, cfg.DSN, log)
	if err != nil {
		return fmt.Errorf("an error occured when init db, err: %w", err)
	}

	wg.Add(1)
	go func() {
		defer log.Info("DB connection were closed")
		defer wg.Done()
		<-ctx.Done()

		db.Close()
	}()

	log.Info("database is connected")

	gkServer, err := server.InitServer(db, db, log, cfg)
	if err != nil {
		componentsErrs <- fmt.Errorf("an occured error when init server, err: %w", err)
	}
	go func(srv *server.GKServer, errs chan<- error) {
		if err := srv.ListenAndServe(); err != nil {
			errs <- fmt.Errorf("listen and serve has failed: %w", err)
		}
	}(gkServer, componentsErrs)

	log.Info("server is running")

	wg.Add(1)
	go func(srv *server.GKServer) {
		defer log.Error("server has been shutdown")
		defer wg.Done()
		<-ctx.Done()

		shutdownTimeoutCtx, cancelShutdownTimeoutCtx := context.WithTimeout(context.Background(), timeoutServerShutdown)
		defer cancelShutdownTimeoutCtx()
		srv.Shutdown(shutdownTimeoutCtx)
	}(gkServer)

	defer func() {
		wg.Wait()
	}()

	select {
	case <-ctx.Done():
	case err := <-componentsErrs:
		log.Error(err.Error())
		cancelCtx()
	}

	go func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), timeoutShutdown)
		defer cancelCtx()

		<-ctx.Done()
		log.Error("failed to gracefully shutdown the service")
	}()
	return nil
}
