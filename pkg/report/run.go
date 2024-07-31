package report

import (
	"github.com/mattfenwick/kubectl-plugins/pkg/utils"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

func Run() {
	kube, err := NewKubernetesForContext("")
	utils.DoOrDie(err)
	rawNsWatcher, err := kube.WatchNamespaces()
	utils.DoOrDie(err)
	nsWatcher := &TypedWatcher[*v1.Namespace]{Name: "namespaces", WatchInterface: rawNsWatcher}
	stop := make(chan struct{})

	go func() {
		HandleNamespaceEvents(nsWatcher)
	}()

	<-stop
	utils.DoOrDie(errors.Errorf("TODO"))
}
