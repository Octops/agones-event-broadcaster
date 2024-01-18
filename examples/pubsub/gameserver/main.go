package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	v1 "agones.dev/agones/pkg/apis/agones/v1"
	"github.com/Octops/agones-event-broadcaster/pkg/broadcaster"
	"github.com/Octops/agones-event-broadcaster/pkg/brokers/pubsub"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"k8s.io/client-go/tools/clientcmd"
)

/*
Make sure you have the following environment variables set when testing the broadcaster using the Pub/Sub broker
KUBECONFIG: valid path to a Kubernetes config file. It can point to a local or remote cluster
PUBSUB_CREDENTIALS: path to the json key file from a service account that has access to Pub/Sub topics
PUBSUB_PROJECT_ID: Google Cloud projectID.

Before running this application the topics must be present on Google Cloud Pub/Sub. This example uses the following topics:
gameserver.events.added: destination of OnAdd events
gameserver.events.updated: destination of OnUpdate events
gameserver.events.deleted: destination of OnDelete events
*/
func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Info("starting application")

	cfg, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))

	opts := option.WithCredentialsFile(os.Getenv("PUBSUB_CREDENTIALS"))
	broker, err := pubsub.NewPubSubBroker(&pubsub.Config{
		ProjectID:       os.Getenv("PUBSUB_PROJECT_ID"),
		OnAddTopicID:    "gameserver.events.added",
		OnUpdateTopicID: "gameserver.events.updated",
		OnDeleteTopicID: "gameserver.events.deleted",
	}, opts)
	if err != nil {
		logrus.WithError(err).Fatal("error creating broker")
	}
	optsbr := &broadcaster.Config{
		SyncPeriod:             15*time.Second,
		ServerPort:             8088,
		MetricsBindAddress:     "0.0.0.0:8095",
		MaxConcurrentReconcile: 5,
		HealthProbeBindAddress: "0.0.0.0:8099",
	}
	gsBroadcaster := broadcaster.New(cfg, broker, optsbr)
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
