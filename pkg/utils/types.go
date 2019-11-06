package utils

import (
	clientV1alpha1 "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// EngineDetails struct is for collecting all the engine-related details
type EngineDetails struct {
	Name         string
	Experiments  []string
	AppLabel     string
	SvcAccount   string
	AppKind      string
	AppNamespace string
	Config       *rest.Config
}

// ExperimentDetails is for collecting all the experiment-related details
type ExperimentDetails struct {
	Env       map[string]string
	ExpLabels map[string]string
	ExpImage  string
	ExpArgs   []string
	JobName   string
	ExpName   string
}

// ClientSets is a collection of clientSets needed
type ClientSets struct {
	KubeClient   *kubernetes.Clientset
	LitmusClient *clientV1alpha1.Clientset
}
