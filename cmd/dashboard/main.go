package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	restorev1alpha1 "github.com/kymaroshq/kymaros/api/v1alpha1"
	kymarosapi "github.com/kymaroshq/kymaros/internal/api"
	kymaroslicense "github.com/kymaroshq/kymaros/internal/license"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(restorev1alpha1.AddToScheme(scheme))
}

func main() {
	var port int
	var staticDir string

	flag.IntVar(&port, "port", 8080, "API server port")
	flag.StringVar(&staticDir, "static-dir", "./dist", "Path to frontend static files")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg := ctrl.GetConfigOrDie()
	c, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		slog.Error("failed to create K8s client", "error", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	licMgr := kymaroslicense.NewManager(c)
	licMgr.Refresh(ctx)
	slog.Info("license loaded", "tier", licMgr.Tier())

	// Periodically refresh license to detect Secret changes and expiration.
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				licMgr.Refresh(ctx)
			}
		}
	}()

	srv := kymarosapi.NewServer(c, port, staticDir, licMgr)

	slog.Info("starting Kymaros dashboard", "port", port, "staticDir", staticDir)
	if err := srv.Start(ctx); err != nil {
		slog.Error("dashboard server failed", "error", err)
		os.Exit(1)
	}
}
