package utils

import (
	"flag"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	clientV1alpha1 "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned"
	volume "github.com/litmuschaos/elves/kubernetes/volume/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/litmuschaos/chaos-runner/pkg/utils/k8s"
	"github.com/litmuschaos/chaos-runner/pkg/utils/litmus"
)

// EngineDetails struct is for collecting all the engine-related details
type EngineDetails struct {
	Name             string
	Experiments      []string
	AppLabel         string
	AppNs            string
	SvcAccount       string
	AppKind          string
	ClientUUID       string
	AuxiliaryAppInfo string
	UID              string
	EngineNamespace  string
	AnnotationKey    string
	AnnotationCheck  string
}

// ExperimentDetails is for collecting all the experiment-related details
type ExperimentDetails struct {
	Name               string
	envMap             map[string]corev1.EnvVar
	ExpLabels          map[string]string
	ExpImage           string
	ExpImagePullPolicy corev1.PullPolicy
	ExpArgs            []string
	JobName            string
	Namespace          string
	ConfigMaps         []v1alpha1.ConfigMap
	Secrets            []v1alpha1.Secret
	HostFileVolumes    []v1alpha1.HostFile
	VolumeOpts         VolumeOpts
	SvcAccount         string
	Annotations        map[string]string
	NodeSelector       map[string]string
	Tolerations        []corev1.Toleration
	SecurityContext    v1alpha1.SecurityContext
	HostPID            bool
	// InstanceID is passed as env inside chaosengine
	// It is separately specified here because this attribute is common for all experiment.
	InstanceID                    string
	ResourceRequirements          v1.ResourceRequirements
	ImagePullSecrets              []corev1.LocalObjectReference
	StatusCheckTimeout            int
	TerminationGracePeriodSeconds int64
}

//VolumeOpts is a strcuture for all volume related operations
type VolumeOpts struct {
	VolumeMounts   []corev1.VolumeMount
	VolumeBuilders []*volume.Builder
}

// ClientSets is a collection of clientSets needed
type ClientSets struct {
	KubeClient   kubernetes.Interface
	LitmusClient clientV1alpha1.Interface
}

// EventAttributes is for collecting all the events-related details
type EventAttributes struct {
	Reason  string
	Message string
	Type    string
	Name    string
}

var (
	// DefaultExpImagePullPolicy contains the defaults value (Always) of imagePullPolicy for exp container
	DefaultExpImagePullPolicy corev1.PullPolicy = "Always"
)

const (
	ExperimentDependencyCheckReason          string = "ExperimentDependencyCheck"
	ExperimentJobCreateReason                string = "ExperimentJobCreate"
	ExperimentJobCleanUpReason               string = "ExperimentJobCleanUp"
	ExperimentSkippedReason                  string = "ExperimentSkipped"
	ExperimentEnvParseErrorReason            string = "EnvParseError"
	ExperimentNotFoundErrorReason            string = "ExperimentNotFound"
	ExperimentJobCreationErrorReason         string = "JobCreationError"
	ExperimentChaosContainerWatchErrorReason string = "ChaosContainerWatchNotPermitted"
	ChaosResourceNotFoundReason              string = "ChaosResourceNotFound"
)

// GenerateClientSetFromKubeConfig will generation both ClientSets (k8s, and Litmus)
func (clientSets *ClientSets) GenerateClientSetFromKubeConfig() error {
	config, err := getKubeConfig()
	if err != nil {
		return err
	}
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

// getKubeConfig setup the config for access cluster resource
func getKubeConfig() (*rest.Config, error) {
	kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()
	// Use in-cluster config if kubeconfig path is specified
	if *kubeconfig == "" {
		config, err := rest.InClusterConfig()
		if err != nil {
			return config, err
		}
	}
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return config, err
	}
	return config, err
}
