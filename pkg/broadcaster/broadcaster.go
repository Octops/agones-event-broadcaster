package broadcaster

import (
	"github.com/Octops/gameserver-events-broadcaster/pkg/controller"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Broadcaster struct {
	logger     *logrus.Entry
	controller *controller.GameServerController
}

func New(logger *logrus.Entry, config *rest.Config) (*Broadcaster, error) {
	gsController, err := controller.NewGameServerController(logger, config)
	if err != nil {
		return nil, err
	}

	return &Broadcaster{
		logger:     logger,
		controller: gsController,
	}, nil
}

func (b *Broadcaster) Start() error {
	if err := b.controller.Run(ctrl.SetupSignalHandler()); err != nil {
		b.logger.WithError(err).Error("broadcaster could not start")
		return err
	}

	return nil
}
