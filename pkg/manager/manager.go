package manager

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

type Options struct {
	SyncPeriod         *time.Duration
	ServerPort         int
	MetricsBindAddress string
}

type Manager struct {
	manager.Manager
}

func New(config *rest.Config, options Options) (*Manager, error) {
	mgr, err := manager.New(config, manager.Options{
		Cache: cache.Options{
			SyncPeriod: options.SyncPeriod,
		},
		Metrics: server.Options{
			BindAddress: options.MetricsBindAddress,
		},
	})

	if err != nil {
		return nil, errors.Wrap(err, "manager could not be created")
	}

	return &Manager{mgr}, nil
}

func (m *Manager) Start(ctx context.Context) error {
	log.SetLogger(zap.New())
	return m.Manager.Start(ctx)
}
