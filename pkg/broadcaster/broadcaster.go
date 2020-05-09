package broadcaster

import (
	"github.com/Octops/gameserver-events-broadcaster/pkg/controller"
	"github.com/Octops/gameserver-events-broadcaster/pkg/runtime/log"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Broadcaster struct {
	logger     *logrus.Entry
	controller *controller.GameServerController
}

func New(config *rest.Config) (*Broadcaster, error) {
	logger := log.NewLoggerWithField("source", "broadcaster")
	gsController, err := controller.NewGameServerController(config)
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
