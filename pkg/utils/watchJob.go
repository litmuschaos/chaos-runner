package utils

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	"github.com/litmuschaos/chaos-runner/pkg/log"
	"github.com/litmuschaos/litmus-go/pkg/utils/retry"
)

var err error

// checkStatusListForExp loops over all the status patched in chaosEngine, to get the one, which has to be updated
// Can go with updated the last status(status[n-1])
// But would'nt work for the parallel execution
func checkStatusListForExp(status []v1alpha1.ExperimentStatuses, ExperimentName string) int {
	for i := range status {
		if status[i].Name == ExperimentName {
			return i
		}
	}
	return -1
}

// GetChaosEngine returns chaosEngine Object
func (engineDetails EngineDetails) GetChaosEngine(clients ClientSets) (*v1alpha1.ChaosEngine, error) {
	var engine *v1alpha1.ChaosEngine
	if err := retry.
		Times(uint(180)).
		Wait(time.Duration(2)).
		Try(func(attempt uint) error {
			engine, err = clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engineDetails.EngineNamespace).Get(engineDetails.Name, metav1.GetOptions{})
			if err != nil {
				return errors.Errorf("unable to get ChaosEngine name: %v, in namespace: %v, error: %v", engineDetails.Name, engineDetails.EngineNamespace, err)
			}
			return nil
		}); err != nil {
		return nil, err
	}
	return engine, nil
}

// PatchChaosEngineStatus updates ChaosEngine with Experiment Status
func (expStatus *ExperimentStatus) PatchChaosEngineStatus(engineDetails EngineDetails, clients ClientSets) error {

	expEngine, err := engineDetails.GetChaosEngine(clients)
	if err != nil {
		return err
	}
	experimentIndex := checkStatusListForExp(expEngine.Status.Experiments, expStatus.Name)
	if experimentIndex == -1 {
		return errors.Errorf("unable to find the status for Experiment: %v in ChaosEngine: %v", expStatus.Name, expEngine.Name)
	}
	expEngine.Status.Experiments[experimentIndex] = v1alpha1.ExperimentStatuses(*expStatus)
	if _, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engineDetails.EngineNamespace).Update(expEngine); err != nil {
		return err
	}
	return nil
}

// GetResultName returns the resultName using the experimentName and engine Name
func GetResultName(engineName, experimentName, instanceID string) string {
	resultName := engineName + "-" + experimentName
	if instanceID != "" {
		resultName = resultName + "-" + instanceID
	}

	return resultName
}

// GetChaosResult returns ChaosResult object.
func (expDetails *ExperimentDetails) GetChaosResult(engineDetails EngineDetails, clients ClientSets) (*v1alpha1.ChaosResult, error) {

	resultName := GetResultName(engineDetails.Name, expDetails.Name, expDetails.InstanceID)
	expResult, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosResults(engineDetails.EngineNamespace).Get(resultName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Errorf("unable to get ChaosResult name: %v in namespace: %v, error: %v", resultName, engineDetails.EngineNamespace, err)
	}
	return expResult, nil
}

// UpdateEngineWithResult will update the result in chaosEngine
// And will delete job if jobCleanUpPolicy is set to "delete"
func (engineDetails EngineDetails) UpdateEngineWithResult(experiment *ExperimentDetails, clients ClientSets) error {
	// Getting the Experiment Result Name
	chaosResult, err := experiment.GetChaosResult(engineDetails, clients)
	if err != nil {
		return err
	}

	var currExpStatus ExperimentStatus
	chaosPod, err := GetChaosPod(experiment, clients)
	if err != nil {
		return errors.Errorf("unable to get the chaos pod, error: %v", err)
	}
	currExpStatus.CompletedExperimentStatus(chaosResult, engineDetails.Name, chaosPod.Name)
	if err = currExpStatus.PatchChaosEngineStatus(engineDetails, clients); err != nil {
		return err
	}

	return nil
}

// DeleteJobAccordingToJobCleanUpPolicy deletes the chaosExperiment Job according to jobCleanUpPolicy
func (engineDetails EngineDetails) DeleteJobAccordingToJobCleanUpPolicy(experiment *ExperimentDetails, clients ClientSets) (v1alpha1.CleanUpPolicy, error) {

	expEngine, err := engineDetails.GetChaosEngine(clients)
	if err != nil {
		return "", err
	}

	switch expEngine.Spec.JobCleanUpPolicy {
	case v1alpha1.CleanUpPolicyDelete:
		log.Infof("deleting the job as jobCleanPolicy is set to %s", expEngine.Spec.JobCleanUpPolicy)
		deletePolicy := metav1.DeletePropagationForeground
		if deleteJobErr := clients.KubeClient.BatchV1().Jobs(experiment.Namespace).Delete(experiment.JobName, &metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		}); deleteJobErr != nil {
			return "", errors.Errorf("unable to delete ChaosExperiment Job name: %v, in namespace: %v, error: %v", experiment.JobName, experiment.Namespace, deleteJobErr)
		}
		log.Infof("%v job is deleted successfully", experiment.JobName)
	case v1alpha1.CleanUpPolicyRetain, "":
		log.Infof("[skip]: skipping the job deletion as jobCleanUpPolicy is set to {%s}", expEngine.Spec.JobCleanUpPolicy)
	default:
		return expEngine.Spec.JobCleanUpPolicy, fmt.Errorf("%s jobCleanUpPolicy not supported", expEngine.Spec.JobCleanUpPolicy)
	}
	return expEngine.Spec.JobCleanUpPolicy, nil
}
