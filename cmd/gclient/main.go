package main

import (
	"context"
	"fmt"

	"github.com/ArtemShalinFe/gophkeeper/internal/build"
	"github.com/ArtemShalinFe/gophkeeper/internal/client"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	log, err := zap.NewProduction()
	if err != nil {
		zap.L().Info("an error occured while init logger err: %w ", zap.Error(err))
		return
	}

	log.Info(fmt.Sprintf("%s", build.NewBuild()))
	client.NewApp(log).Start(ctx)
}
