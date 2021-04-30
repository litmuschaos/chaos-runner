package k8s

import (
	"github.com/pkg/errors"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// GenerateK8sClientSet will generation k8s client
func GenerateK8sClientSet(config *rest.Config) (*kubernetes.Clientset, error) {
	k8sClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Errorf("unable to generate kubernetes clientSet, error: %v: ", err)
	}
	return k8sClientSet, nil
}
