package utils

import (
	"flag"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	clientV1alpha1 "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned"
	volume "github.com/litmuschaos/elves/kubernetes/volume/v1alpha1"
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
	envMap             map[string]v1.EnvVar
	ExpLabels          map[string]string
	ExpImage           string
	ExpImagePullPolicy v1.PullPolicy
	ExpArgs            []string
	ExpCommand         []string
	JobName            string
	Namespace          string
	ConfigMaps         []v1alpha1.ConfigMap
	Secrets            []v1alpha1.Secret
	HostFileVolumes    []v1alpha1.HostFile
	VolumeOpts         VolumeOpts
	SvcAccount         string
	Annotations        map[string]string
	NodeSelector       map[string]string
	Tolerations        []v1.Toleration
	SecurityContext    v1alpha1.SecurityContext
	HostPID            bool
	// InstanceID is passed as env inside chaosengine
	// It is separately specified here because this attribute is common for all experiment.
	InstanceID                    string
	ResourceRequirements          v1.ResourceRequirements
	ImagePullSecrets              []v1.LocalObjectReference
	StatusCheckTimeout            int
	TerminationGracePeriodSeconds int64
	DefaultAppHealthCheck         string
}

//VolumeOpts is a strcuture for all volume related operations
type VolumeOpts struct {
	VolumeMounts   []v1.VolumeMount
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
	DefaultExpImagePullPolicy v1.PullPolicy = "Always"
)

const (
	// ExperimentDependencyCheckReason contains the reason for the dependency check event
	ExperimentDependencyCheckReason string = "ExperimentDependencyCheck"
	// ExperimentJobCreateReason contains the reason for the job creation event
	ExperimentJobCreateReason string = "ExperimentJobCreate"
	// ExperimentJobCleanUpReason contains the reason for the job cleanup event
	ExperimentJobCleanUpReason string = "ExperimentJobCleanUp"
	// ExperimentSkippedReason contains the reason for the experiment skip event
	ExperimentSkippedReason string = "ExperimentSkipped"
	// ExperimentEnvParseErrorReason contains the reason for the env-parse-error event
	ExperimentEnvParseErrorReason string = "EnvParseError"
	// ExperimentNotFoundErrorReason contains the reason for the experiment-not-found event
	ExperimentNotFoundErrorReason string = "ExperimentNotFound"
	// ExperimentJobCreationErrorReason contains the reason for the job-creation-error event
	ExperimentJobCreationErrorReason string = "JobCreationError"
	// ExperimentChaosContainerWatchErrorReason contains the reason for the watch-job-error event
	ExperimentChaosContainerWatchErrorReason string = "ChaosContainerWatchNotPermitted"
	// ChaosResourceNotFoundReason contains the reason for the chaos-resources-not-found event
	ChaosResourceNotFoundReason string = "ChaosResourceNotFound"
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
	return clientcmd.BuildConfigFromFlags("", *kubeconfig)
}
