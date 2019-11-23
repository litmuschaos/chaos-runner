package utils

import (
	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// checkStatusListForExp loops over all the status patched in chaosEngine, to get the one, which has to be updated
// Can go with updated the last status(status[n-1])
// But would'nt work for the pararllel execution
<<<<<<< HEAD
func checkStatusListForExp(status []v1alpha1.ExperimentStatuses, jobName string) int {
=======
func checkStatusListForJob(status []v1alpha1.ExperimentStatuses, jobName string) int {
>>>>>>> 7fc58d356d7f488f50b0af0134e6d881b469225b
	for i := range status {
		if status[i].Name == jobName {
			return i
		}
	}
	return -1
}
<<<<<<< HEAD
=======
func checkStatusListForExp(status []v1alpha1.ExperimentStatuses, expName string) int {
	for i := range status {
		if status[i].Name == expName {
			return i
		}
	}
	return -1
}
>>>>>>> 7fc58d356d7f488f50b0af0134e6d881b469225b

// WatchingJobtillCompletion will watch the JOb, and update it's status
func WatchingJobtillCompletion(perExperiment ExperimentDetails, engineDetails EngineDetails, clients ClientSets) error {
	var jobStatus int32
	jobStatus = 1
	// jobStatus will remain 1, if its running
	// So, is used to loop over the check for its completion
	for jobStatus == 1 {
		expEngine, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engineDetails.AppNamespace).Get(engineDetails.Name, metav1.GetOptions{})
		if err != nil {
			log.Print(err)
			return err
		}
		var currExpStatus v1alpha1.ExperimentStatuses
		currExpStatus.Name = perExperiment.JobName
		currExpStatus.Status = "Running"
		currExpStatus.LastUpdateTime = metav1.Now()
		currExpStatus.Verdict = "Waiting For Completion"
<<<<<<< HEAD
		checkForjobName := checkStatusListForExp(expEngine.Status.Experiments, perExperiment.JobName)
		if checkForjobName == -1 {
			expEngine.Status.Experiments = append(expEngine.Status.Experiments, currExpStatus)
=======
		checkForjobName := checkStatusListForJob(expEngine.Status.Experiments, perExperiment.JobName)
		if checkForjobName == -1 {
			if checkStatusListForExp(expEngine.Status.Experiments, perExperiment.ExpName) == -1 {
				log.Fatal("Not able to find any Experiment with this name")
			} else {
				expEngine.Status.Experiments[checkStatusListForExp(expEngine.Status.Experiments, perExperiment.ExpName)] = currExpStatus
			}
>>>>>>> 7fc58d356d7f488f50b0af0134e6d881b469225b
		} else {
			expEngine.Status.Experiments[checkForjobName].LastUpdateTime = metav1.Now()
		}
		log.Info("Patching Engine")
		_, updateErr := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engineDetails.AppNamespace).Update(expEngine)
		if updateErr != nil {
			return err
		}
		getJob, err := clients.KubeClient.BatchV1().Jobs(engineDetails.AppNamespace).Get(perExperiment.JobName, metav1.GetOptions{})
		if err != nil {
			log.Infoln("Unable to get the job : ", err)
			return err
		}
		jobStatus = getJob.Status.Active
		log.Infoln("Watching for Job Name : "+perExperiment.JobName+" status of Job : ", jobStatus)
		//log.Infoln(jobStatus)
		time.Sleep(5 * time.Second)
	}
	return nil

}

// GetResultName returns the resultName using the experimentName and engine Name
func GetResultName(engineDetails EngineDetails, i int) string {
	resultName := engineDetails.Name + "-" + engineDetails.Experiments[i]
	log.Info("ResultName : " + resultName)
	return resultName
}

// UpdateResultWithJobAndDeletingJob will update hte resutl in chaosEngine
// And will delete job if jobCleanUpPolicy is set to "delete"
func UpdateResultWithJobAndDeletingJob(engineDetails EngineDetails, clients ClientSets, resultName string, perExperiment ExperimentDetails) error {
	expResult, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosResults(engineDetails.AppNamespace).Get(resultName, metav1.GetOptions{})
	if err != nil {
		log.Infoln("Unable to get chaosResult Resource")
		log.Panic(err)
		return err
	}
	verdict := expResult.Spec.ExperimentStatus.Verdict
<<<<<<< HEAD
	phase := expResult.Spec.ExperimentStatus.Phase
=======
	//phase := expResult.Spec.ExperimentStatus.Phase
>>>>>>> 7fc58d356d7f488f50b0af0134e6d881b469225b
	expEngine, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engineDetails.AppNamespace).Get(engineDetails.Name, metav1.GetOptions{})
	if err != nil {
		log.Print(err)
		return err
	}
	log.Info(expEngine)
	var currExpStatus v1alpha1.ExperimentStatuses
	currExpStatus.Name = perExperiment.JobName
<<<<<<< HEAD
	currExpStatus.Status = phase
	currExpStatus.LastUpdateTime = metav1.Now()
	currExpStatus.Verdict = verdict
	checkForjobName := checkStatusListForExp(expEngine.Status.Experiments, perExperiment.JobName)
=======
	currExpStatus.Status = "Completed"
	currExpStatus.LastUpdateTime = metav1.Now()
	currExpStatus.Verdict = verdict
	checkForjobName := checkStatusListForJob(expEngine.Status.Experiments, perExperiment.JobName)
>>>>>>> 7fc58d356d7f488f50b0af0134e6d881b469225b
	if checkForjobName == -1 {
		expEngine.Status.Experiments = append(expEngine.Status.Experiments, currExpStatus)
	} else {
		expEngine.Status.Experiments[checkForjobName] = currExpStatus
	}
	log.Info(expEngine)
	_, updateErr := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engineDetails.AppNamespace).Update(expEngine)
	if updateErr != nil {
		log.Info("Updating Resource Error : ", updateErr)
		return updateErr
	}
	if expEngine.Spec.JobCleanUpPolicy == "delete" {
		log.Infoln("Will delete the job as jobCleanPolicy is set to : " + expEngine.Spec.JobCleanUpPolicy)
		deleteJob := clients.KubeClient.BatchV1().Jobs(engineDetails.AppNamespace).Delete(perExperiment.JobName, &metav1.DeleteOptions{})
		if deleteJob != nil {
			log.Panic(deleteJob)
			return deleteJob
		}

	}
	return nil
}
