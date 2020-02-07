package litmus

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"

	clientV1alpha1 "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned"
)

// GenerateLitmusClientSet will generate a LitmusClient
func GenerateLitmusClientSet(config *rest.Config) (*clientV1alpha1.Clientset, error) {
	litmusClientSet, err := clientV1alpha1.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create LitmusClientSet: %v", err)
	}
	return litmusClientSet, nil
}
