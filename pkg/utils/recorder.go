package utils

import (
	litmuschaosScheme "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned/scheme"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
)

func generateEventRecorder(kubeClient *kubernetes.Clientset) (record.EventRecorder, error) {
	err := litmuschaosScheme.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, err
	}
	eventBroadcaster := record.NewBroadcaster()
	//"github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned/typed/litmuschaos/v1alpha1"
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "chaos-runner"})
	return recorder, nil
}

// NewEventRecorder initalizes EventRecorder with Resource as ChaosEngine
func NewEventRecorder(clients ClientSets, engineDetails EngineDetails) (*Recorder, error) {
	engineForEvent, err := engineDetails.GetChaosEngine(clients)
	if err != nil {
		return &Recorder{}, err
	}
	eventBroadCaster, err := generateEventRecorder(clients.KubeClient)
	if err != nil {
		return &Recorder{}, err
	}
	return &Recorder{
		EventRecorder: eventBroadCaster,
		EventResource: engineForEvent,
	}, nil
}

const (
	experimentDependencyCheck string = "ExperimentDependencyCheck"
	experimentJobCreate       string = "ExperimentJobCreate"
	experimentJobCleanUp      string = "ExperimentJobCleanUp"
	experimentSkipped         string = "ExperimentSkipped"
)

// ExperimentDepedencyCheck is an standard event spawned just after validating
// experiment dependent resources such as ChaosExperiment, ConfigMaps and Secrets.
func (r Recorder) ExperimentDepedencyCheck(experimentName string) {
	r.EventRecorder.Eventf(r.EventResource, corev1.EventTypeNormal, experimentDependencyCheck, "Experiment resources validated for Chaos Experiment: '%v'", experimentName)
}

// ExperimentJobCreate is an standard event spawned just after
// starting chaosExperiment Job
func (r Recorder) ExperimentJobCreate(experimentName string, jobName string) {
	r.EventRecorder.Eventf(r.EventResource, corev1.EventTypeNormal, experimentJobCreate, "Experiment Job '%v' created for Chaos Experiment '%v'", jobName, experimentName)
}

// ExperimentJobCleanUp is an standard event spawned just after
// starting ChaosExperiment Job
func (r Recorder) ExperimentJobCleanUp(experimentName string, jobCleanUpPolicy string) {
	if jobCleanUpPolicy == "delete" {
		r.EventRecorder.Eventf(r.EventResource, corev1.EventTypeNormal, experimentJobCleanUp, "Experiment Job for Chaos Experiment '%v' will be deleted", experimentName)
	} else {
		r.EventRecorder.Eventf(r.EventResource, corev1.EventTypeNormal, experimentJobCleanUp, "Experiment Job for Chaos Experiment '%v' will be retained", experimentName)
	}
}

// ExperimentSkipped is an standard event spawned just after
// an experiment is skipped
func (r Recorder) ExperimentSkipped(experimentName string) {
	r.EventRecorder.Eventf(r.EventResource, corev1.EventTypeNormal, experimentSkipped, "Experiment Job creation failed, skipping Chaos Experiment: '%v'", experimentName)
}
