package utils

import (
	"fmt"
	"time"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func GetJobStatus(experimentDetails *ExperimentDetails, clients ClientSets) int32 {

	getJob, err := clients.KubeClient.BatchV1().Jobs(experimentDetails.Namespace).Get(experimentDetails.JobName, metav1.GetOptions{})
	if err != nil {
		log.Infoln("Unable to get the job : ", err)
		return -1
	}
	jobStatus := getJob.Status.Active
	log.Infof("Watchin the Job: %v, Status of Job: %v", experimentDetails.JobName, jobStatus)
	return jobStatus
}

func (engineDetails EngineDetails) GetChaosEngine(clients ClientSets) (*v1alpha1.ChaosEngine, error) {
	expEngine, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engineDetails.AppNamespace).Get(engineDetails.Name, metav1.GetOptions{})
	if err != nil {
		log.Infof("Unable to get chaosEngine Name: %v, in NameSpace: %v", engineDetails.Name, engineDetails.AppNamespace)
		return nil, err
	}
	return expEngine, nil
}
func (expStatus *ExperimentStatus) RunningExperimentStatus(experimentDetails *ExperimentDetails) {
	expStatus.Name = experimentDetails.JobName
	expStatus.Status = "Running"
	expStatus.Verdict = "Awaited"
	expStatus.LastUpdateTime = metav1.Now()
}

func (expStatus *ExperimentStatus) PatchChaosEngineStatus(engineDetails EngineDetails, clients ClientSets) error {

	//log.Infof("Printing the Experiment Status: %v", expStatus)
	expEngine, err := engineDetails.GetChaosEngine(clients)
	if err != nil {
		return err
	}
	jobIndex := checkStatusListForExp(expEngine.Status.Experiments, expStatus.Name)
	//log.Infof("Printing the Experiment Index: %v", jobIndex)
	if jobIndex == -1 {
		log.Infof("Unable to find the status for JobName: %v in ChaosEngine: %v", expStatus.Name, expEngine.Name)
		return fmt.Errorf("Unable to find the chaosJob in ChaosEngine")
	}
	expEngine.Status.Experiments[jobIndex] = v1alpha1.ExperimentStatuses(*expStatus)
	_, updateErr := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engineDetails.AppNamespace).Update(expEngine)
	if updateErr != nil {
		return err
	}
	return nil
}
func (engineDetails EngineDetails) WatchingJobtillCompletion(experiment *ExperimentDetails, clients ClientSets) error {

	jobStatus := GetJobStatus(experiment, clients)
	if jobStatus == -1 {
		return fmt.Errorf("Unable to get the chaosExperiment Job Status")
	}
	for jobStatus == 1 {
		//checkForjobName := checkStatusListForExp(expEngine.Status.Experiments, experiment.JobName)
		var expStatus ExperimentStatus
		expStatus.RunningExperimentStatus(experiment)
		expStatus.PatchChaosEngineStatus(engineDetails, clients)
		time.Sleep(5 * time.Second)
		jobStatus = GetJobStatus(experiment, clients)

	}
	return nil

}

// WatchingJobtillCompletion will watch the JOb, and update it's status
/*func WatchingJobtillCompletion(experiment *ExperimentDetails, engineDetails EngineDetails, clients ClientSets) error {
	var jobStatus int32
	jobStatus = 1
	// jobStatus will remain 1, if its running
	// So, is used to loop over the check for its completion
	for jobStatus == 1 {

		expEngine, err := engineDetails.GetChaosEngine(clients)
		if err != nil {
			return err
		}

		var currExpStatus v1alpha1.ExperimentStatuses
		currExpStatus.Name = experiment.JobName
		currExpStatus.Status = "Running"
		currExpStatus.LastUpdateTime = metav1.Now()
		currExpStatus.Verdict = "Waiting For Completion"
		checkForjobName := checkStatusListForExp(expEngine.Status.Experiments, experiment.JobName)
		if checkForjobName == -1 {
			expEngine.Status.Experiments = append(expEngine.Status.Experiments, currExpStatus)
		} else {
			expEngine.Status.Experiments[checkForjobName].LastUpdateTime = metav1.Now()
		}
		log.Info("Patching Engine")
		_, updateErr := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engineDetails.AppNamespace).Update(expEngine)
		if updateErr != nil {
			return err
		}
		jobStatus = GetJobStatus(experiment, clients)
		if jobStatus == -1 {
			log.Infof("Unable to get JobStatus")
		}
		//log.Infoln(jobStatus)
		time.Sleep(5 * time.Second)
	}
	return nil

}
*/
// GetResultName returns the resultName using the experimentName and engine Name
func GetResultName(engineName string, experimentName string) string {
	resultName := engineName + "-" + experimentName
	log.Info("ResultName : " + resultName)
	return resultName
}
func (experimentDetails *ExperimentDetails) GetChaosResult(engineDetails EngineDetails, clients ClientSets) (*v1alpha1.ChaosResult, error) {

	resultName := GetResultName(engineDetails.Name, experimentDetails.Name)
	expResult, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosResults(engineDetails.AppNamespace).Get(resultName, metav1.GetOptions{})
	if err != nil {
		log.Infoln("Unable to get chaosResult Resource")
		log.Infoln(err)
		return nil, err
	}
	return expResult, nil
}

func (expStatus *ExperimentStatus) FinalExperimentStatus(chaosResult *v1alpha1.ChaosResult, experimentDetails *ExperimentDetails) {
	//var currExpStatus v1alpha1.ExperimentStatuses
	expStatus.Name = experimentDetails.JobName
	expStatus.Status = "Execultion Successful"
	expStatus.LastUpdateTime = metav1.Now()
	expStatus.Verdict = chaosResult.Spec.ExperimentStatus.Verdict
	//return currExpStatus
}

// UpdateResultWithJobAndDeletingJob will update hte resutl in chaosEngine
// And will delete job if jobCleanUpPolicy is set to "delete"
func (engineDetails EngineDetails) UpdateResultWithJobAndDeletingJob(experiment *ExperimentDetails, clients ClientSets) error {
	// Getting the Experiment Result Name
	chaosResult, err := experiment.GetChaosResult(engineDetails, clients)
	if err != nil {
		return err
	}
	expEngine, err := engineDetails.GetChaosEngine(clients)
	if err != nil {
		return err
	}

	var currExpStatus ExperimentStatus
	currExpStatus.FinalExperimentStatus(chaosResult, experiment)

	currExpStatus.PatchChaosEngineStatus(engineDetails, clients)

	if expEngine.Spec.JobCleanUpPolicy == "delete" {
		log.Infoln("Will delete the job as jobCleanPolicy is set to : " + expEngine.Spec.JobCleanUpPolicy)
		deleteJob := clients.KubeClient.BatchV1().Jobs(engineDetails.AppNamespace).Delete(experiment.JobName, &metav1.DeleteOptions{})
		if deleteJob != nil {
			log.Infoln(deleteJob)
			return deleteJob
		}

	}
	return nil
}
