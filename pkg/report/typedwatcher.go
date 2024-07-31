package report

import (
	"context"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/watch"
)

type TypedEvent[A any] struct {
	Type   watch.EventType
	Object A
	Error  interface{}
}

type TypedWatcher[A any] struct {
	Name           string
	WatchInterface watch.Interface
}

func (w *TypedWatcher[A]) GetEvents(ctx context.Context) <-chan *TypedEvent[A] {
	raw := w.WatchInterface.ResultChan()
	typed := make(chan *TypedEvent[A])
	go func() {
		for {
			select {
			case e := <-raw:
				logrus.Debugf("kube watcher %s found event of type %s", w.Name, e.Type)
				switch e.Type {
				case watch.Error:
					typed <- &TypedEvent[A]{Type: e.Type, Error: e.Object}
				default:
					typed <- &TypedEvent[A]{Type: e.Type, Object: e.Object.(A)}
				}
			case <-ctx.Done():
				logrus.Infof("exiting watcher %s", w.Name)
				w.WatchInterface.Stop()
				return
			}
		}
	}()
	return typed
}
