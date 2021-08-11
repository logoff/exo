// Separate logd service for testing in isolation. Unused for production
// deployments.  By default, binds a unix domain socket for easy discovery from
// exop.

package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/deref/exo/internal/config"
	josh "github.com/deref/exo/internal/josh/server"
	"github.com/deref/exo/internal/logd"
	"github.com/deref/exo/internal/logd/api"
	"github.com/deref/exo/internal/telemetry"
	"github.com/deref/exo/internal/util/cmdutil"
	"github.com/deref/exo/internal/util/logging"
)

func main() {
	ctx := context.Background()

	logger := logging.Default()
	ctx = logging.ContextWithLogger(ctx, logger)

	tel := telemetry.New(ctx, &config.TelemetryConfig{
		Disable: true,
	})
	ctx = telemetry.ContextWithTelemetry(ctx, tel)

	ctx, done := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer done()

	cfg := &config.Config{}
	config.MustLoadDefault(cfg)
	paths := cmdutil.MustMakeDirectories(cfg)

	logd := &logd.Service{
		VarDir:     paths.VarDir,
		SyslogPort: cfg.Log.SyslogPort,
		Logger:     logger,
	}
	logd.Debug = true

	{
		ctx, shutdown := context.WithCancel(ctx)
		defer shutdown()
		go func() {
			if err := logd.Run(ctx); err != nil {
				cmdutil.Warnf("log collector error: %w", err)
			}
		}()
	}

	var network, addr string
	if cfg.HTTPPort == 0 {
		network = "unix"
		addr = filepath.Join(paths.VarDir, "logd.sock")
		_ = os.Remove(addr)
	} else {
		network = "tcp"
		addr = fmt.Sprintf(":%d", cfg.HTTPPort)
	}
	listener, err := net.Listen(network, addr)
	if err != nil {
		cmdutil.Fatalf("error listening: %v", err)
	}
	fmt.Println("listening at", addr)

	muxb := josh.NewMuxBuilder("/")
	api.BuildLogCollectorMux(muxb, func(req *http.Request) api.LogCollector {
		return &logd.LogCollector
	})
	cmdutil.Serve(ctx, listener, &http.Server{
		Handler: muxb.Build(),
	})
}
