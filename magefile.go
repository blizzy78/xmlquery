//go:build mage
// +build mage

package main

import (
	"context"
	"fmt"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var Default = Build

func Lint() error {
	if err := sh.Run("go", "run", "github.com/golangci/golangci-lint/cmd/golangci-lint@latest", "run", "-c", ".golangci.yml"); err != nil {
		return fmt.Errorf("go run golangci-lint: %w", err)
	}

	if err := sh.Run("go", "run", "github.com/blizzy78/consistent/cmd/consistent@latest", "."); err != nil {
		return fmt.Errorf("go run consistent: %w", err)
	}

	return nil
}

func Build(ctx context.Context) error {
	mg.SerialCtxDeps(ctx, Lint)

	if err := sh.Run("go", "build", "-o", "xmlquery", "."); err != nil {
		return fmt.Errorf("go build: %w", err)
	}

	return nil
}
