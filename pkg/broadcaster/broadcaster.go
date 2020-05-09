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

	gsBroadcaster := &Broadcaster{
		logger: logger,
	}

	gsController, err := controller.NewGameServerController(config, gsBroadcaster)
	if err != nil {
		return nil, err
	}

	gsBroadcaster.controller = gsController

	return gsBroadcaster, nil
}

func (b *Broadcaster) Start() error {
	if err := b.controller.Run(ctrl.SetupSignalHandler()); err != nil {
		b.logger.WithError(err).Error("broadcaster could not start")
		return err
	}

	return nil
}

func (b *Broadcaster) OnAdd(obj interface{}) error {
	b.logger.Debug(obj)

	return nil
}

func (b *Broadcaster) OnUpdate(oldObj interface{}, newObj interface{}) error {
	b.logger.Debug(oldObj, newObj)

	return nil
}

func (b *Broadcaster) OnDelete(obj interface{}) error {
	b.logger.Debug(obj)

	return nil
}
