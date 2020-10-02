package main

import (
	v1 "agones.dev/agones/pkg/apis/agones/v1"
	"context"
	"flag"
	"github.com/Octops/agones-event-broadcaster/pkg/broadcaster"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	gsBroadcaster := broadcaster.New(cfg, broker, 15*time.Second)
	gsBroadcaster.WithWatcherFor(&v1.GameServer{})
	if err := gsBroadcaster.Build(); err != nil {
		logrus.WithError(err).Fatal("error creating broadcaster")
	}

	if err := gsBroadcaster.Start(); err != nil {
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
