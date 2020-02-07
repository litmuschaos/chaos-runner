package utils

import (
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
)

// checkStatusListForExp loops over all the status patched in chaosEngine, to get the one, which has to be updated
// Can go with updated the last status(status[n-1])
// But would'nt work for the pararllel execution
func checkStatusListForExp(status []v1alpha1.ExperimentStatuses, jobName string) int {
	for i := range status {
		if status[i].Name == jobName {
			return i
		}
	}
	return -1
}

// GetJobStatus gets status of the job
func GetJobStatus(experimentDetails *ExperimentDetails, clients ClientSets) (int32, error) {

	getJob, err := clients.KubeClient.BatchV1().Jobs(experimentDetails.Namespace).Get(experimentDetails.JobName, metav1.GetOptions{})
	if err != nil {
		//TODO: check for jobStatus should not return -1 directly, look for best practices.
		return -1, errors.Wrapf(err, "Unable to get ChaosExperiment Job, due to error: %v", err)
	}
	//TODO:check the container of the Job, rather than going with the JobStatus.
	jobStatus := getJob.Status.Active
	return jobStatus, nil
}

// GetChaosEngine returns chaosEngine Object
func (engineDetails EngineDetails) GetChaosEngine(clients ClientSets) (*v1alpha1.ChaosEngine, error) {
	expEngine, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engineDetails.AppNamespace).Get(engineDetails.Name, metav1.GetOptions{})
	if err != nil {

		return nil, errors.Wrapf(err, "Unable to get ChaosEngine Name: %v, in namespace: %v, due to error: %v", engineDetails.Name, engineDetails.AppNamespace, err)
	}
	return expEngine, nil
}

// PatchChaosEngineStatus updates ChaosEngine with Experiment Status
func (expStatus *ExperimentStatus) PatchChaosEngineStatus(engineDetails EngineDetails, clients ClientSets) error {

	expEngine, err := engineDetails.GetChaosEngine(clients)
	if err != nil {
		return err
	}
	jobIndex := checkStatusListForExp(expEngine.Status.Experiments, expStatus.Name)
	if jobIndex == -1 {
		return errors.Wrapf(err, "Unable to find the status for JobName: %v in ChaosEngine: %v", expStatus.Name, expEngine.Name)
	}
	expEngine.Status.Experiments[jobIndex] = v1alpha1.ExperimentStatuses(*expStatus)
	if _, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engineDetails.AppNamespace).Update(expEngine); err != nil {
		return err
	}
	return nil
}

// WatchJobForCompletion watches the chaosExperiment job for completions
func (engineDetails EngineDetails) WatchJobForCompletion(experiment *ExperimentDetails, clients ClientSets) error {

	//TODO: use watch rather than checking for status manually.
	jobStatus := int32(1)
	var err error
	for jobStatus == 1 {
		jobStatus, err = GetJobStatus(experiment, clients)
		if err != nil {
			return err
		}
		//checkForjobName := checkStatusListForExp(expEngine.Status.Experiments, experiment.JobName)
		var expStatus ExperimentStatus
		expStatus.AwaitedExperimentStatus(experiment)
		if err := expStatus.PatchChaosEngineStatus(engineDetails, clients); err != nil {
			return errors.Wrapf(err, "Unable to patch ChaosEngine in namespace: %v, due to error: %v", engineDetails.AppNamespace, err)
		}
		time.Sleep(5 * time.Second)

	}
	return nil
}

// GetResultName returns the resultName using the experimentName and engine Name
func GetResultName(engineName string, experimentName string) string {
	resultName := engineName + "-" + experimentName
	return resultName
}

// GetChaosResult returns ChaosResult object.
func (experimentDetails *ExperimentDetails) GetChaosResult(engineDetails EngineDetails, clients ClientSets) (*v1alpha1.ChaosResult, error) {

	resultName := GetResultName(engineDetails.Name, experimentDetails.Name)
	expResult, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosResults(engineDetails.AppNamespace).Get(resultName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to get ChaosResult Name: %v in namespace: %v, due to error: %v", resultName, engineDetails.AppNamespace, err)
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
	currExpStatus.CompletedExperimentStatus(chaosResult, experiment)
	if err = currExpStatus.PatchChaosEngineStatus(engineDetails, clients); err != nil {
		return err
	}

	return nil
}

// DeleteJobAccordingToJobCleanUpPolicy deletes the chaosExperiment Job according to jobCleanUpPolicy
func (engineDetails EngineDetails) DeleteJobAccordingToJobCleanUpPolicy(experiment *ExperimentDetails, clients ClientSets) error {

	expEngine, err := engineDetails.GetChaosEngine(clients)
	if err != nil {
		return err
	}

	if expEngine.Spec.JobCleanUpPolicy == "delete" {
		klog.V(0).Infoln("Will delete the job as jobCleanPolicy is set to : " + expEngine.Spec.JobCleanUpPolicy)

		deletePolicy := metav1.DeletePropagationForeground
		deleteJob := clients.KubeClient.BatchV1().Jobs(engineDetails.AppNamespace).Delete(experiment.JobName, &metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		})
		if deleteJob != nil {
			return errors.Wrapf(err, "Unable to delete ChaosExperiment Job Name: %v, in namespace: %v, due to error: %v", experiment.JobName, experiment.Namespace, err)
		}
	}
	return nil
}
