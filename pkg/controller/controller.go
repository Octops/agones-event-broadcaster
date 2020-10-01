package controller

import "time"

type Options struct {
	SyncPeriod time.Duration
}

type BroadcasterController interface {
	Run(stop <-chan struct{}) error
}
