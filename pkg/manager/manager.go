package manager

import (
	"context"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"
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
		SyncPeriod:         options.SyncPeriod,
		Port:               options.ServerPort,
		MetricsBindAddress: options.MetricsBindAddress,
	})

	if err != nil {
		return nil, errors.Wrap(err, "manager could not be created")
	}

	return &Manager{mgr}, nil
}

func (m *Manager) Start(ctx context.Context) error {
	return m.Manager.Start(ctx)
}
