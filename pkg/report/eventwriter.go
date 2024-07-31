package report

import (
	"context"
	"fmt"

	"github.com/mattfenwick/kubectl-plugins/pkg/utils"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func HandleNamespaceEvents(ctx context.Context, nsWatcher *TypedWatcher[*v1.Namespace]) {
	for e := range nsWatcher.GetEvents(ctx) {
		switch e.Type {
		case watch.Added:
			fmt.Printf("add event -- %+v\n", e.Object.Name)
		case watch.Bookmark:
			fmt.Printf("don't care about bookmark event\n")
		case watch.Deleted:
			fmt.Printf("delete event -- %s\n", e.Object.Name)
		case watch.Modified:
			fmt.Printf("modify event -- %s\n", e.Object.Name)
		case watch.Error:
			fmt.Printf("error event -- %+v\n", e.Object)
		default:
			utils.DoOrDie(errors.Errorf("invalid type, %s", e.Type))
		}
	}
}

func HandlePodEvents(ctx context.Context, nsWatcher *TypedWatcher[*v1.Pod]) {
	for e := range nsWatcher.GetEvents(ctx) {
		switch e.Type {
		case watch.Added:
			fmt.Printf("add pod event -- %s/%s\n", e.Object.Namespace, e.Object.Name)
		case watch.Bookmark:
			fmt.Printf("don't care about pod bookmark event\n")
		case watch.Deleted:
			fmt.Printf("delete pod event -- %s/%s\n", e.Object.Namespace, e.Object.Name)
		case watch.Modified:
			fmt.Printf("modify pod event -- %s/%s\n", e.Object.Namespace, e.Object.Name)
		case watch.Error:
			fmt.Printf("error pod event -- %+v\n", e.Object)
		default:
			utils.DoOrDie(errors.Errorf("invalid type, %s", e.Type))
		}
	}
}

func HandleConfigMapEvents(ctx context.Context, nsWatcher *TypedWatcher[*v1.ConfigMap]) {
	for e := range nsWatcher.GetEvents(ctx) {
		switch e.Type {
		case watch.Added:
			fmt.Printf("add cm event -- %s/%s, \n  %s\n", e.Object.Namespace, e.Object.Name, e.Object.Data)
		case watch.Bookmark:
			fmt.Printf("don't care about cm bookmark event\n")
		case watch.Deleted:
			fmt.Printf("delete cm event -- %s/%s\n", e.Object.Namespace, e.Object.Name)
		case watch.Modified:
			fmt.Printf("modify cm event -- %s/%s\n", e.Object.Namespace, e.Object.Name)
		case watch.Error:
			fmt.Printf("error cm event -- %+v\n", e.Object)
		default:
			utils.DoOrDie(errors.Errorf("invalid type, %s", e.Type))
		}
	}
}

func HandleSecretEvents(ctx context.Context, nsWatcher *TypedWatcher[*v1.Secret]) {
	for e := range nsWatcher.GetEvents(ctx) {
		switch e.Type {
		case watch.Added:
			fmt.Printf("add secret event -- %s/%s, \n  %s\n", e.Object.Namespace, e.Object.Name, e.Object.Data)
		case watch.Bookmark:
			fmt.Printf("don't care about secret bookmark event\n")
		case watch.Deleted:
			fmt.Printf("delete secret event -- %s/%s\n", e.Object.Namespace, e.Object.Name)
		case watch.Modified:
			fmt.Printf("modify secret event -- %s/%s\n", e.Object.Namespace, e.Object.Name)
		case watch.Error:
			fmt.Printf("error secret event -- %+v\n", e.Object)
		default:
			utils.DoOrDie(errors.Errorf("invalid type, %s", e.Type))
		}
	}
}

func HandleEventEvents(ctx context.Context, nsWatcher *TypedWatcher[*v1.Event]) {
	for e := range nsWatcher.GetEvents(ctx) {
		switch e.Type {
		case watch.Added:
			fmt.Printf("add event event -- %s/%s, %s %s %s\n", e.Object.Namespace, e.Object.Name, e.Object.Message, e.Object.Reason, &e.Object.Source)
		case watch.Bookmark:
			fmt.Printf("don't care about event bookmark event\n")
		case watch.Deleted:
			fmt.Printf("delete event event -- %s/%s\n", e.Object.Namespace, e.Object.Name)
		case watch.Modified:
			fmt.Printf("modify event event -- %s/%s\n", e.Object.Namespace, e.Object.Name)
		case watch.Error:
			fmt.Printf("error event event -- %+v\n", e.Object)
		default:
			utils.DoOrDie(errors.Errorf("invalid type, %s", e.Type))
		}
	}
}
