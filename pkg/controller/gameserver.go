package controller

import (
	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"context"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"log"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type GameServerController struct {
	logger *logrus.Entry
	manager.Manager
}

func NewGameServerController(logger *logrus.Entry, config *rest.Config) (*GameServerController, error) {
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

func (c *GameServerController) Run(stop <-chan struct{}) error {
	if err := c.Start(stop); err != nil {
		c.logger.WithError(err).Error("error starting controller manager")
		return err
	}

	return nil
}

type reconciler struct {
	client.Client
	scheme *runtime.Scheme
}

func (r *reconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()

	gameServer := &agonesv1.GameServer{}
	if err := r.Get(ctx, req.NamespacedName, gameServer); err != nil {
		if apierrors.IsNotFound(err) {
			log.Println(err)
			return ctrl.Result{}, nil
		}

		log.Println("Error:", err.Error())
		return reconcile.Result{}, err
	}

	log.Println(req.NamespacedName, gameServer.Status.State)

	return reconcile.Result{}, nil
}
