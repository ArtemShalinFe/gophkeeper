package main

import (
	"context"

	"github.com/ArtemShalinFe/gophkeeper/internal/client"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	client.NewApp(zap.L()).Start(ctx)
}
