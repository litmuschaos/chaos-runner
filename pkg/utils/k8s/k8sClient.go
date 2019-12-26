package k8s

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// GenerateK8sClientSet will generation k8s client
func GenerateK8sClientSet(config *rest.Config) (*kubernetes.Clientset, error) {
	k8sClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to generate kubernetes clientSet %s: ", err)
	}
	return k8sClientSet, nil
}
