package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/Tomap-Tomap/GophKeeper/tui"
	"github.com/Tomap-Tomap/GophKeeper/tui/buildinfo"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	if err := tui.Run(ctx, buildinfo.New(buildVersion, buildDate, buildCommit)); err != nil {
		fmt.Printf("run tui2: %v", err)
	}
}
