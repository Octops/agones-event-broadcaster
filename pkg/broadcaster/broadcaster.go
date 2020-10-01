package broadcaster

import (
	"github.com/Octops/agones-event-broadcaster/pkg/brokers"
	"github.com/Octops/agones-event-broadcaster/pkg/controller"
	"github.com/Octops/agones-event-broadcaster/pkg/events"
	"github.com/Octops/agones-event-broadcaster/pkg/runtime/log"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

// Broadcaster receives events (Add, Update and Delete) sent by the controller
// and uses a Broker to publish those events.
type Broadcaster struct {
	logger     *logrus.Entry
	controller *controller.GameServerController
	brokers.Broker
}

// New returns a new GameServer broadcaster
// It required a config to be passed to the GameServer controller
// and a broker that will be publishing messages
func New(config *rest.Config, broker brokers.Broker, syncPeriod time.Duration) (*Broadcaster, error) {
	logger := log.NewLoggerWithField("source", "broadcaster")

	gsBroadcaster := &Broadcaster{
		logger: logger,
		Broker: broker,
	}

	gsController, err := controller.NewGameServerController(config, gsBroadcaster, controller.Options{SyncPeriod: syncPeriod})
	if err != nil {
		return nil, err
	}

	gsBroadcaster.controller = gsController

	return gsBroadcaster, nil
}

// Start run the controller that sends events back to the broadcaster event handlers
func (b *Broadcaster) Start() error {
	if err := b.controller.Run(ctrl.SetupSignalHandler()); err != nil {
		b.logger.WithError(err).Error("broadcaster could not start")
		return err
	}

	return nil
}

// OnAdd is the event handler that reacts to Add events
func (b *Broadcaster) OnAdd(obj interface{}) error {
	if b.Broker == nil {
		b.logger.Warn("broker is not available for the broadcaster, message will not be published")
		return nil
	}

	message := &events.EventMessage{
		Body: obj,
	}

	event := events.GameServerAdded(message)

	return b.Publish(event)
}

// OnUpdate is the event handler that reacts to Update events
func (b *Broadcaster) OnUpdate(oldObj interface{}, newObj interface{}) error {
	if b.Broker == nil {
		b.logger.Warn("a broker is not available for the broadcaster, message will not be published")
		return nil
	}

	body := struct {
		OldObj interface{}
		NewObj interface{}
	}{
		OldObj: oldObj,
		NewObj: newObj,
	}

	message := &events.EventMessage{
		Body: body,
	}

	event := events.GameServerUpdated(message)

	return b.Publish(event)
}

// OnDelete is the event handler that reacts to Delete events
func (b *Broadcaster) OnDelete(obj interface{}) error {
	if b.Broker == nil {
		b.logger.Warn("a broker is not available for the broadcaster, message will not be published")
		return nil
	}

	message := &events.EventMessage{
		Body: obj,
	}

	event := events.GameServerDeleted(message)

	return b.Publish(event)
}

// Publish will publish the event wrapped on a envelope using the broker available
func (b *Broadcaster) Publish(event events.Event) error {
	envelope, err := b.Broker.BuildEnvelope(event)
	if err != nil {
		b.logger.WithError(err).Error("error building envelope")
		return err
	}

	if err = b.Broker.SendMessage(envelope); err != nil {
		b.logger.WithError(err).Error("error sending envelope")
		return err
	}

	return nil
}
