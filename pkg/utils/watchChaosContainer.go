package utils

import (
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetChaosContainerStatus gets status of the chaos container
func GetChaosContainerStatus(experimentDetails *ExperimentDetails, clients ClientSets) (bool, error) {

	isCompleted := false
	PodList, err := clients.KubeClient.CoreV1().Pods(experimentDetails.Namespace).List(metav1.ListOptions{LabelSelector: "job-name=" + experimentDetails.JobName})
	if err != nil || len(PodList.Items) == 0 {
		return false, errors.Wrapf(err, "Unable to get the chaos pod, due to error: %v", err)
	}

	for _, pod := range PodList.Items {
		if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodSucceeded {
			for _, container := range pod.Status.ContainerStatuses {

				//NOTE: The name of container inside chaos-pod is same as the chaos job name
				// we only have one container inside chaos pod to inject the chaos
				// looking the chaos container is completed or not
				if container.Name == experimentDetails.JobName && container.State.Terminated != nil {
					if container.State.Terminated.Reason == "Completed" {
						isCompleted = !container.Ready
					}

				}
			}
		}
	}
	return isCompleted, nil
}

// WatchChaosContainerForCompletion watches the chaos container for completion
func (engineDetails EngineDetails) WatchChaosContainerForCompletion(experiment *ExperimentDetails, clients ClientSets) error {

	//TODO: use watch rather than checking for status manually.
	isChaosCompleted := false
	var err error
	for !isChaosCompleted {
		isChaosCompleted, err = GetChaosContainerStatus(experiment, clients)
		if err != nil {
			return err
		}

		var expStatus ExperimentStatus
		expStatus.AwaitedExperimentStatus(experiment)
		if err := expStatus.PatchChaosEngineStatus(engineDetails, clients); err != nil {
			return errors.Wrapf(err, "Unable to patch ChaosEngine in namespace: %v, due to error: %v", engineDetails.EngineNamespace, err)
		}
		time.Sleep(5 * time.Second)

	}
	return nil
}
