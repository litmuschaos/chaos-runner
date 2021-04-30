package utils

import (
	"time"

	"github.com/litmuschaos/litmus-go/pkg/utils/retry"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetChaosPod gets the chaos experiment pod object launched by the runner
func GetChaosPod(expDetails *ExperimentDetails, clients ClientSets) (*corev1.Pod, error) {
	var chaosPodList *corev1.PodList
	var err error

	delay := 2
	err = retry.
		Times(uint(expDetails.StatusCheckTimeout / delay)).
		Wait(time.Duration(delay) * time.Second).
		Try(func(attempt uint) error {
			chaosPodList, err = clients.KubeClient.CoreV1().Pods(expDetails.Namespace).List(metav1.ListOptions{LabelSelector: "job-name=" + expDetails.JobName})
			if err != nil || len(chaosPodList.Items) == 0 {
				return errors.Errorf("unable to get the chaos pod, error: %v", err)
			} else if len(chaosPodList.Items) > 1 {
				// Cases where experiment pod is rescheduled by the job controller due to
				// issues while the older pod is still not cleaned-up
				return errors.Errorf("Multiple pods exist with same job-name label")
			}
			return nil
		})
	if err != nil {
		return nil, err
	}

	// Note: We error out upon existence of multiple exp pods for the same experiment
	// & hence use index [0]
	chaosPod := &chaosPodList.Items[0]
	return chaosPod, nil
}

// GetChaosContainerStatus gets status of the chaos container
func GetChaosContainerStatus(experimentDetails *ExperimentDetails, clients ClientSets) (bool, error) {

	isCompleted := false

	pod, err := GetChaosPod(experimentDetails, clients)
	if err != nil {
		return false, errors.Errorf("unable to get the chaos pod, error: %v", err)
	}
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

	} else if pod.Status.Phase == corev1.PodPending {
		delay := 2
		err := retry.
			Times(uint(experimentDetails.StatusCheckTimeout / delay)).
			Wait(time.Duration(delay) * time.Second).
			Try(func(attempt uint) error {
				pod, err := GetChaosPod(experimentDetails, clients)
				if err != nil {
					return errors.Errorf("unable to get the chaos pod, error: %v", err)
				}
				if pod.Status.Phase == corev1.PodPending {
					return errors.Errorf("chaos pod is in %v state", corev1.PodPending)
				}
				return nil
			})
		if err != nil {
			return isCompleted, err
		}
	} else if pod.Status.Phase == corev1.PodFailed {
		return isCompleted, errors.Errorf("status check failed as chaos pod status is %v", pod.Status.Phase)
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
		chaosPod, err := GetChaosPod(experiment, clients)
		if err != nil {
			return errors.Errorf("unable to get the chaos pod, error: %v", err)
		}

		expStatus.AwaitedExperimentStatus(experiment.Name, engineDetails.Name, chaosPod.Name)
		if err := expStatus.PatchChaosEngineStatus(engineDetails, clients); err != nil {
			return errors.Errorf("unable to patch ChaosEngine in namespace: %v, error: %v", engineDetails.EngineNamespace, err)
		}
		time.Sleep(2 * time.Second)
	}
	return nil
}
