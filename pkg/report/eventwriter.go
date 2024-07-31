package report

import (
	"context"
	"fmt"

	"github.com/mattfenwick/kubectl-plugins/pkg/utils"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func HandleNamespaceEvents(nsWatcher *TypedWatcher[*v1.Namespace]) {
	for e := range nsWatcher.GetEvents(context.TODO()) {
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
		fmt.Printf("Namespace event! %T\n", e.Object)
	}
}
