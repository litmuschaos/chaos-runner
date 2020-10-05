package utils

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	"github.com/litmuschaos/chaos-runner/pkg/log"
)

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
	expEngine, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engineDetails.EngineNamespace).Get(engineDetails.Name, metav1.GetOptions{})
	if err != nil {

		return nil, errors.Errorf("Unable to get ChaosEngine Name: %v, in namespace: %v, error: %v", engineDetails.Name, engineDetails.EngineNamespace, err)
	}
	return expEngine, nil
}

// PatchChaosEngineStatus updates ChaosEngine with Experiment Status
func (expStatus *ExperimentStatus) PatchChaosEngineStatus(engineDetails EngineDetails, clients ClientSets) error {

	expEngine, err := engineDetails.GetChaosEngine(clients)
	if err != nil {
		return err
	}
	experimentIndex := checkStatusListForExp(expEngine.Status.Experiments, expStatus.Name)
	if experimentIndex == -1 {
		return errors.Errorf("Unable to find the status for Experiment: %v in ChaosEngine: %v", expStatus.Name, expEngine.Name)
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
func (experimentDetails *ExperimentDetails) GetChaosResult(engineDetails EngineDetails, clients ClientSets) (*v1alpha1.ChaosResult, error) {

	resultName := GetResultName(engineDetails.Name, experimentDetails.Name, experimentDetails.InstanceID)
	expResult, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosResults(engineDetails.EngineNamespace).Get(resultName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Errorf("Unable to get ChaosResult Name: %v in namespace: %v, error: %v", resultName, engineDetails.EngineNamespace, err)
	}
	return expResult, nil
}

// UpdateEngineWithResult will update hte resutl in chaosEngine
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
		return errors.Errorf("Unable to get the chaos pod, error: %v", err)
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

	if expEngine.Spec.JobCleanUpPolicy == v1alpha1.CleanUpPolicyDelete || string(expEngine.Spec.JobCleanUpPolicy) == "" {
		log.Infof("deleting the job as jobCleanPolicy is set to %v", expEngine.Spec.JobCleanUpPolicy)

		deletePolicy := metav1.DeletePropagationForeground
		deleteJob := clients.KubeClient.BatchV1().Jobs(experiment.Namespace).Delete(experiment.JobName, &metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		})
		if deleteJob != nil {
			return "", errors.Errorf("Unable to delete ChaosExperiment Job Name: %v, in namespace: %v, error: %v", experiment.JobName, experiment.Namespace, err)
		}
	}
	return expEngine.Spec.JobCleanUpPolicy, nil
}
