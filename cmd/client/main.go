package main

import "go.uber.org/zap"

func main() {
	if err := run(); err != nil {
		zap.S().Fatalf("an occured fatal err: %w", err)
	}
}

func run() error {
	return nil
}
