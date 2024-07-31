package report

import (
	"context"

	"github.com/mattfenwick/kubectl-plugins/pkg/utils"
	"github.com/pkg/errors"
)

func Run() {
	kube, err := NewKubernetesForContext("")
	utils.DoOrDie(err)

	rootCtx := context.Background()

	nsWatcher, err := kube.WatchNamespaces()
	utils.DoOrDie(err)
	podWatcher, err := kube.WatchPods()
	utils.DoOrDie(err)
	configMapWatcher, err := kube.WatchConfigMaps()
	utils.DoOrDie(err)
	secretWatcher, err := kube.WatchSecrets()
	utils.DoOrDie(err)
	eventWatcher, err := kube.WatchEvents()
	utils.DoOrDie(err)

	stop := make(chan struct{})

	go func() {
		HandleNamespaceEvents(rootCtx, nsWatcher)
	}()
	go func() {
		HandlePodEvents(rootCtx, podWatcher)
	}()
	go func() {
		HandleConfigMapEvents(rootCtx, configMapWatcher)
	}()
	go func() {
		HandleSecretEvents(rootCtx, secretWatcher)
	}()
	go func() {
		HandleEventEvents(rootCtx, eventWatcher)
	}()

	<-stop
	utils.DoOrDie(errors.Errorf("TODO"))
}
