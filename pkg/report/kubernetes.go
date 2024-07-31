package report

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Kubernetes struct {
	ClientSet  *kubernetes.Clientset
	RestConfig *rest.Config
}

func NewKubernetesForContext(context string) (*Kubernetes, error) {
	logrus.Debugf("instantiating k8s Clientset for context %s", context)
	kubeConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: context}).ClientConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to build config")
	}
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to instantiate Clientset")
	}
	return &Kubernetes{
		ClientSet:  clientset,
		RestConfig: kubeConfig,
	}, nil
}

func (k *Kubernetes) WatchNamespaces() (*TypedWatcher[*v1.Namespace], error) {
	raw, err := k.ClientSet.CoreV1().Namespaces().Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get namespaces watcher")
	}
	return &TypedWatcher[*v1.Namespace]{Name: "namespaces", WatchInterface: raw}, nil
}

func (k *Kubernetes) WatchPods() (*TypedWatcher[*v1.Pod], error) {
	raw, err := k.ClientSet.CoreV1().Pods("").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get pods watcher")
	}
	return &TypedWatcher[*v1.Pod]{Name: "pods", WatchInterface: raw}, nil
}

func (k *Kubernetes) WatchConfigMaps() (*TypedWatcher[*v1.ConfigMap], error) {
	raw, err := k.ClientSet.CoreV1().ConfigMaps("").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get configmaps watcher")
	}
	return &TypedWatcher[*v1.ConfigMap]{Name: "configmaps", WatchInterface: raw}, nil
}

func (k *Kubernetes) WatchSecrets() (*TypedWatcher[*v1.Secret], error) {
	raw, err := k.ClientSet.CoreV1().Secrets("").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get secrets watcher")
	}
	return &TypedWatcher[*v1.Secret]{Name: "secrets", WatchInterface: raw}, nil
}

func (k *Kubernetes) WatchEvents() (*TypedWatcher[*v1.Event], error) {
	raw, err := k.ClientSet.CoreV1().Events("").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get event watcher")
	}
	return &TypedWatcher[*v1.Event]{Name: "events", WatchInterface: raw}, nil
}
