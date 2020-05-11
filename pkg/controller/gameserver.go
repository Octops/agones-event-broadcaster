package controller

import (
	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"context"
	"fmt"
	"github.com/Octops/gameserver-events-broadcaster/pkg/events/handlers"
	"github.com/Octops/gameserver-events-broadcaster/pkg/runtime/log"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// GameServerController watches Agones GameServer events
// and notify the event handlers
type GameServerController struct {
	logger *logrus.Entry
	manager.Manager
}

// reconciler is notified every time an event happens.
// It can differentiate between events types.
// The GameServer controller uses the eventHandler for a more grained control.
// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.6.0/pkg/reconcile?tab=doc#Reconciler
type reconciler struct {
	client.Client
	scheme *runtime.Scheme
}

// NewGameServerController returns a GameServer controller that uses the informed eventHandler
// to notify the Broadcaster about reconcile events for Agones GameServers
func NewGameServerController(config *rest.Config, eventHandler handlers.EventHandler) (*GameServerController, error) {
	logger := log.NewLoggerWithField("source", "GameServerController")
	mgr, err := manager.New(config, manager.Options{})
	if err != nil {
		return nil, err
	}

	err = ctrl.NewControllerManagedBy(mgr).
		For(&agonesv1.GameServer{}).
		Owns(&corev1.Pod{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				// Implement some logic here and if returns true if you think that
				// this event should be sent to the reconciler or false otherwise
				return true
			},
			DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
				return true
			},
			UpdateFunc: func(updateEvent event.UpdateEvent) bool {
				return true
			},
			GenericFunc: func(genericEvent event.GenericEvent) bool {
				return true
			},
		}).
		Watches(&source.Kind{Type: &agonesv1.GameServer{}}, &handler.Funcs{
			CreateFunc: func(createEvent event.CreateEvent, limitingInterface workqueue.RateLimitingInterface) {
				// OnAdd is triggered only when the controller is syncing its cache.
				// It does not map ot the resource creation event triggered by Kubernetes
				request := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Namespace: createEvent.Meta.GetNamespace(),
						Name:      createEvent.Meta.GetName(),
					},
				}

				defer limitingInterface.Done(request)

				if err := eventHandler.OnAdd(createEvent.Object); err != nil {
					limitingInterface.AddRateLimited(request)
					return
				}

				limitingInterface.Forget(request)
			},
			UpdateFunc: func(updateEvent event.UpdateEvent, limitingInterface workqueue.RateLimitingInterface) {
				request := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Namespace: updateEvent.MetaNew.GetNamespace(),
						Name:      updateEvent.MetaNew.GetName(),
					},
				}

				defer limitingInterface.Done(request)

				if err := eventHandler.OnUpdate(updateEvent.ObjectOld, updateEvent.ObjectNew); err != nil {
					limitingInterface.AddRateLimited(request)
					return
				}

				limitingInterface.Forget(request)
			},
			DeleteFunc: func(deleteEvent event.DeleteEvent, limitingInterface workqueue.RateLimitingInterface) {

				request := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Namespace: deleteEvent.Meta.GetNamespace(),
						Name:      deleteEvent.Meta.GetName(),
					},
				}

				if err := eventHandler.OnDelete(deleteEvent.Object); err != nil {
					limitingInterface.AddRateLimited(request)
					return
				}

				limitingInterface.Forget(request)
			},
		}).
		Complete(&reconciler{
			Client: mgr.GetClient(),
			scheme: mgr.GetScheme(),
		})

	if err != nil {
		return nil, err
	}

	controller := &GameServerController{
		logger:  logger,
		Manager: mgr,
	}

	return controller, nil
}

// Run starts the GameServerController and watches reconcile events for Agones GameServers
func (c *GameServerController) Run(stop <-chan struct{}) error {
	if err := c.Start(stop); err != nil {
		c.logger.WithError(err).Error("error starting controller manager")
		return err
	}

	return nil
}

// Reconcile is called on every reconcile event. It does not differ between add, update, delete.
// Its function is purely informative and events are handled back to the broadcaster specific event handlers.
func (r *reconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()

	gameServer := &agonesv1.GameServer{}
	if err := r.Get(ctx, req.NamespacedName, gameServer); err != nil {
		if apierrors.IsNotFound(err) {
			logrus.WithError(err).Error()
			return ctrl.Result{}, nil
		}

		logrus.WithError(err).Error()

		return reconcile.Result{}, err
	}

	msg := fmt.Sprintf("OnReconcile: %s - %s", req.NamespacedName, gameServer.Status.State)
	logrus.Debug(msg)

	return reconcile.Result{}, nil
}
