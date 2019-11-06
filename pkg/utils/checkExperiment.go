package utils

import (
	"fmt"
	clientV1alpha1 "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// GenerateClientSets will generation both ClientSets (k8s, and Litmus)
func GenerateClientSets(config *rest.Config) (*kubernetes.Clientset, *clientV1alpha1.Clientset, error) {
	k8sClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate kubernetes clientSet %s: ", err)
	}
	litmusClientSet, err := clientV1alpha1.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate litmus clientSet %s: ", err)
	}
	return k8sClientSet, litmusClientSet, nil
}

// CheckExperimentInAppNamespace will check the experiment in the app namespace
func CheckExperimentInAppNamespace(appns string, chaosExperiment string, config *rest.Config) bool {
	_, litmusClientSet, err := GenerateClientSets(config)
	//fmt.Println(k8sClientSet)
	if err != nil {
		log.Error(err)
	}
	_, err = litmusClientSet.LitmuschaosV1alpha1().ChaosExperiments(appns).Get(chaosExperiment, metav1.GetOptions{})
	isNotFound := k8serror.IsNotFound(err)
	return isNotFound
}
