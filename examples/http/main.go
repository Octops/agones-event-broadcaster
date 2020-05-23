package main

import (
	"flag"
	"github.com/Octops/agones-event-broadcaster/pkg/broadcaster"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
)

var (
	masterURL  string
	kubeconfig string
	address    string
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

	broker := NewHTTPBroker(address)

	http.HandleFunc("/", broker.Handler)
	go func() {
		logrus.Infof("server listening at %s", address)
		logrus.Fatal(http.ListenAndServe(address, nil))
	}()

	gsBroadcaster, err := broadcaster.New(cfg, broker)
	if err != nil {
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
		flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	}

	flag.StringVar(&address, "addr", ":8000", "The address of the HTTP server.")
}
