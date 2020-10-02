package broadcaster

import (
	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"github.com/Octops/agones-event-broadcaster/pkg/brokers"
	"github.com/Octops/agones-event-broadcaster/pkg/controller"
	"github.com/Octops/agones-event-broadcaster/pkg/events"
	"github.com/Octops/agones-event-broadcaster/pkg/runtime/log"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"
)

// Broadcaster receives events (Add, Update and Delete) sent by the controller
// and uses a Broker to publish those events.
type Broadcaster struct {
	logger      *logrus.Entry
	controllers []*controller.AgonesController
	brokers.Broker
	manager.Manager
}

// New returns a new GameServer broadcaster
// It required a config to be passed to the GameServer controller
// and a broker that will be publishing messages
func New(config *rest.Config, broker brokers.Broker, syncPeriod time.Duration) (*Broadcaster, error) {
	logger := log.NewLoggerWithField("source", "broadcaster")

	broadcaster := &Broadcaster{
		logger: logger,
		Broker: broker,
	}

	/*
		Stopped Here
		create controllers based on flags
		boot message
		controller logger fields
	*/
	mgr, err := manager.New(config, manager.Options{
		SyncPeriod: &syncPeriod,
	})

	if err != nil {
		return nil, err
	}

	broadcaster.Manager = mgr

	bcCtrl, err := controller.NewAgonesController(mgr, broadcaster, controller.Options{
		For:  &agonesv1.GameServer{},
		Owns: &corev1.Pod{},
	})
	if err != nil {
		return nil, err
	}

	broadcaster.AddController(bcCtrl)

	fleet, err := controller.NewAgonesController(mgr, broadcaster, controller.Options{
		For:  &agonesv1.Fleet{},
		Owns: &corev1.Pod{},
	})
	if err != nil {
		return nil, err
	}

	broadcaster.AddController(fleet)

	return broadcaster, nil
}

func (b *Broadcaster) AddController(controller *controller.AgonesController) {
	b.controllers = append(b.controllers, controller)
}

// Start run the controller that sends events back to the broadcaster event handlers
func (b *Broadcaster) Start() error {
	chDone := ctrl.SetupSignalHandler()
	if err := b.Manager.Start(chDone); err != nil {
		b.logger.Fatal(errors.Wrap(err, "broadcaster could not start"))
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

	event := events.ForAdded(message)

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

	event := events.ForUpdated(message)

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

	event := events.ForDeleted(message)

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
