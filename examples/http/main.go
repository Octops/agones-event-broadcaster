package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	v1 "agones.dev/agones/pkg/apis/agones/v1"
	"github.com/Octops/agones-event-broadcaster/pkg/broadcaster"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	masterURL  string
	kubeconfig string
	addr       string
)

func main() {
	flag.Parse()

	if kubeconfig == "" {
		kubeconfig = flag.Lookup("kubeconfig").Value.String()
	}

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		logrus.Fatalf("Error building kubeconfig: %s", err.Error())
	}
	cfg.Timeout = time.Minute * 5

	ctx := context.Background()

	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	defer func() {
		signal.Stop(c)
		cancel()
	}()

	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	broker := NewHTTPBroker(addr)
	broker.Start(ctx)
	opts := &broadcaster.Config{
		SyncPeriod:             15*time.Second,
		ServerPort:             8090,
		MetricsBindAddress:     "0.0.0.0:8095",
		MaxConcurrentReconcile: 5,
		HealthProbeBindAddress: "0.0.0.0:8099",
	}
	gsBroadcaster := broadcaster.New(cfg, broker, opts)
	gsBroadcaster.WithWatcherFor(&v1.GameServer{})
	if err := gsBroadcaster.Build(); err != nil {
		logrus.WithError(err).Fatal("error creating broadcaster")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := gsBroadcaster.Start(ctx); err != nil {
		logrus.WithError(err).Fatal("error starting broadcaster")
	}
}

func init() {
	if flag.Lookup("kubeconfig") == nil {
		flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	}

	if flag.Lookup("master") == nil {
		flag.StringVar(&masterURL, "master", "", "The addr of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	}

	flag.StringVar(&addr, "addr", ":8000", "The addr of the HTTP server.")
}
