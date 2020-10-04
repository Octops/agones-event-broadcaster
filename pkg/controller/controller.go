package controller

import (
	"github.com/Octops/agones-event-broadcaster/pkg/events/handlers"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
)

type Options struct {
	For  runtime.Object
	Owns runtime.Object
}

// AgonesController watches for events associated to a particular resource type like GameServers or Fleets.
// It uses the passed EventHandler argument to send back the current state of the world.
type AgonesController struct {
	logger *logrus.Entry
	manager.Manager
}

// Reconciler handles events when resources are reconciled. The interval is configured on the Manager's level.
type Reconciler struct {
	logger *logrus.Entry
	obj    runtime.Object
	client.Client
	scheme *runtime.Scheme
}

func NewAgonesController(mgr manager.Manager, eventHandler handlers.EventHandler, options Options) (*AgonesController, error) {
	optFor := reflect.TypeOf(options.For).Elem().String()
	logger := logrus.WithFields(logrus.Fields{
		"source":          "controller",
		"controller_type": optFor,
	})

	err := ctrl.NewControllerManagedBy(mgr).
		For(options.For).
		//Owns(options.Owns). //TODO: Assigning Owns duplicates the number of reconcile calls.
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

				//TODO: Investigate if controller require this Done. Keeping doubles the reconcile calls
				//defer limitingInterface.Done(request)

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

				//TODO: Investigate if controller require this Done. Keeping doubles the reconcile calls
				//defer limitingInterface.Done(request)

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
			logger: logger,
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

	logger.Infof("controller created for resource of type %s", optFor)
	return controller, nil
}

// Warning: This method is possible not meant to be used. It has a particular use case but the broadcaster uses a shorter
// Sync period that triggers OnUpdate events. Right now this Reconcile function is useless for the broadcaster.
// It should be explored in the future.

// TODO: Evaluate is Reconcile should be made an argument for the Controller. Reconcile can be used for general uses cases
// where control over very specific events matter. Right now it is just a STDOUT output.
// Reconcile is called on every reconcile event. It does not differ between add, update, delete.
// Its function is purely informative and events are handled back to the broadcaster specific event handlers.
func (r *Reconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	//ctx := context.Background()
	//obj := r.obj.DeepCopyObject()
	//if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
	//	if apierrors.IsNotFound(err) {
	//		r.logger.WithField("type", reflect.TypeOf(obj).String()).Debugf("resource \"%s\" not found", req.NamespacedName)
	//		return ctrl.Result{}, nil
	//	}
	//
	//	r.logger.WithError(err).Error()
	//
	//	return reconcile.Result{}, err
	//}

	//r.logger.Debugf("OnReconcile: %s (%s)", req.NamespacedName, reflect.TypeOf(obj).String())

	return reconcile.Result{}, nil
}
