package utils

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	"time"

	litmuschaosScheme "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned/scheme"
)

func generateEventRecorder(kubeClient *kubernetes.Clientset) (record.EventRecorder, error) {
	err := litmuschaosScheme.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, err
	}
	eventBroadcaster := record.NewBroadcaster()
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

// ExperimentDepedencyCheck is an standard event spawned just after validating
// experiment dependent resources such as ChaosExperiment, ConfigMaps and Secrets.
func (r Recorder) ExperimentDepedencyCheck(experimentName string) {
	r.EventRecorder.Eventf(r.EventResource, corev1.EventTypeNormal, ExperimentDependencyCheckReason, "Experiment resources validated for Chaos Experiment: '%s'", experimentName)
	time.Sleep(5 * time.Second)
}

// ExperimentJobCreate is an standard event spawned just after
// starting chaosExperiment Job
func (r Recorder) ExperimentJobCreate(experimentName string, jobName string) {
	r.EventRecorder.Eventf(r.EventResource, corev1.EventTypeNormal, ExperimentJobCreateReason, "Experiment Job '%s' created for Chaos Experiment '%s'", jobName, experimentName)
	time.Sleep(5 * time.Second)
}

// ExperimentJobCleanUp is an standard event spawned just after
// starting ChaosExperiment Job
func (r Recorder) ExperimentJobCleanUp(experiment *ExperimentDetails, jobCleanUpPolicy string) {
	if jobCleanUpPolicy == "delete" {
		r.EventRecorder.Eventf(r.EventResource, corev1.EventTypeNormal, ExperimentJobCleanUpReason, "Experiment Job '%s' is deleted", experiment.JobName)
		time.Sleep(5 * time.Second)
	} else {
		r.EventRecorder.Eventf(r.EventResource, corev1.EventTypeNormal, ExperimentJobCleanUpReason, "Experiment Job '%s' will be retained", experiment.JobName)
		time.Sleep(5 * time.Second)
	}
}

// ExperimentSkipped is an standard event spawned just after
// an experiment is skipped
func (r Recorder) ExperimentSkipped(experimentName string, reason string) {
	r.EventRecorder.Eventf(r.EventResource, corev1.EventTypeWarning, reason, "Experiment Job creation failed, skipping Chaos Experiment: '%s'", experimentName)
	time.Sleep(5 * time.Second)
}
