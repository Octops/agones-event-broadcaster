package controller

import (
	"context"
	"github.com/Octops/agones-event-broadcaster/pkg/events/handlers"
	"github.com/Octops/agones-event-broadcaster/pkg/runtime/log"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

type Options struct {
	SyncPeriod time.Duration
	For        runtime.Object
	Owns       runtime.Object
}

type BroadcasterController interface {
	Run(stop <-chan struct{}) error
}

type AgonesController struct {
	logger *logrus.Entry
	manager.Manager
}

type Reconciler struct {
	obj runtime.Object
	client.Client
	scheme *runtime.Scheme
}

func NewAgonesController(config *rest.Config, eventHandler handlers.EventHandler, options Options) (*AgonesController, error) {
	logger := log.NewLoggerWithField("source", "FleetController")
	mgr, err := manager.New(config, manager.Options{
		SyncPeriod: &options.SyncPeriod,
	})
	if err != nil {
		return nil, err
	}

	err = ctrl.NewControllerManagedBy(mgr).
		For(options.For).
		Owns(options.Owns).
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
		Watches(&source.Kind{Type: options.For}, &handler.Funcs{
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
		Complete(&Reconciler{
			obj:    options.For,
			Client: mgr.GetClient(),
			scheme: mgr.GetScheme(),
		})

	if err != nil {
		return nil, err
	}

	controller := &AgonesController{
		logger:  logger,
		Manager: mgr,
	}

	return controller, nil
}

// Run starts the AgonesController and watches reconcile events for Agones resources
func (c *AgonesController) Run(stop <-chan struct{}) error {
	c.logger.Debug("starting controller")
	if err := c.Start(stop); err != nil {
		c.logger.WithError(err).Error("error starting controller manager")
		return err
	}

	return nil
}

// Reconcile is called on every reconcile event. It does not differ between add, update, delete.
// Its function is purely informative and events are handled back to the broadcaster specific event handlers.
func (r *Reconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()

	obj := r.obj.DeepCopyObject()
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		if apierrors.IsNotFound(err) {
			logrus.WithField("type", reflect.TypeOf(obj).String()).Debugf("resource \"%s\" not found", req.NamespacedName)
			return ctrl.Result{}, nil
		}

		logrus.WithError(err).Error()

		return reconcile.Result{}, err
	}

	logrus.Debugf("OnReconcile: %s - %s", req.NamespacedName, reflect.TypeOf(obj).String())

	return reconcile.Result{}, nil
}
