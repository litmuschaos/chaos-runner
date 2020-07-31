package utils

import (
	"time"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


var (
	//chaosPod holds the experiment pod spec
	chaosPod = &corev1.Pod{}
)


// GetChaosPodName gets the name of the chaos experiment pod launched by the runner
func GetChaosPod(expDetails *ExperimentDetails, clients ClientSets) (*corev1.Pod, error){
	chaosPodList, err := clients.KubeClient.CoreV1().Pods(expDetails.Namespace).List(metav1.ListOptions{LabelSelector: "job-name=" + expDetails.JobName})
	if err != nil || len(chaosPodList.Items) == 0 {
    	return nil, errors.Wrapf(err, "Unable to get the chaos pod, due to error: %v", err)
	} else if len(chaosPodList.Items) > 1 {
		// Cases where experiment pod is rescheduled by the job controller due to 
		// issues while the older pod is still not cleaned-up
		return nil, errors.New("Multiple pods exist with same jobname label")
	}

	for _, pod := range chaosPodList.Items {
		chaosPod = &pod
	}

	return chaosPod, nil
}



// GetChaosContainerStatus gets status of the chaos container
func GetChaosContainerStatus(experimentDetails *ExperimentDetails, clients ClientSets) (bool, error) {

	isCompleted := false

	pod, err := GetChaosPod(experimentDetails, clients)
	if err != nil {
		return false, errors.Wrapf(err, "Unable to get the chaos pod, due to error: %v", err)
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
			return errors.Wrapf(err, "Unable to get the chaos pod, due to error: %v", err)
        }

		expStatus.AwaitedExperimentStatus(experiment.Name, engineDetails.Name, chaosPod.Name)
		if err := expStatus.PatchChaosEngineStatus(engineDetails, clients); err != nil {
			return errors.Wrapf(err, "Unable to patch ChaosEngine in namespace: %v, due to error: %v", engineDetails.EngineNamespace, err)
		}
		time.Sleep(5 * time.Second)

	}
	return nil
}
