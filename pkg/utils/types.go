package utils

import (
	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	clientV1alpha1 "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned"
	volume "github.com/litmuschaos/kube-helper/kubernetes/volume/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/litmuschaos/chaos-executor/pkg/utils/k8s"
	"github.com/litmuschaos/chaos-executor/pkg/utils/litmus"
)

// EngineDetails struct is for collecting all the engine-related details
type EngineDetails struct {
	Name             string
	Experiments      []string
	AppLabel         string
	SvcAccount       string
	AppKind          string
	AppNamespace     string
	ClientUUID       string
	AuxiliaryAppInfo string
}

// ExperimentDetails is for collecting all the experiment-related details
type ExperimentDetails struct {
	Name       string
	Env        map[string]string
	ExpLabels  map[string]string
	ExpImage   string
	ExpArgs    []string
	JobName    string
	Namespace  string
	ConfigMaps []v1alpha1.ConfigMap
	Secrets    []v1alpha1.Secret
	VolumeOpts VolumeOpts
	SvcAccount string
}

//VolumeOpts is a strcuture for all volume related operations
type VolumeOpts struct {
	VolumeMounts   []corev1.VolumeMount
	VolumeBuilders []*volume.Builder
}

// ClientSets is a collection of clientSets needed
type ClientSets struct {
	KubeClient   *kubernetes.Clientset
	LitmusClient *clientV1alpha1.Clientset
}

type LogStruct struct {
	Operation    string
	ResourceType string
	ResourceName string
	Verbosity    int32
	String       string
	Namespace    string
}

// GenerateClientSets will generation both ClientSets (k8s, and Litmus)
func (clientSets *ClientSets) GenerateClientSets(config *rest.Config) error {
	k8sClientSet, err := k8s.GenerateK8sClientSet(config)
	if err != nil {
		return err
	}
	litmusClientSet, err := litmus.GenerateLitmusClientSet(config)
	if err != nil {
		return err
	}
	clientSets.KubeClient = k8sClientSet
	clientSets.LitmusClient = litmusClientSet

	return nil

}
