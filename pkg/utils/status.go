package utils

import (
	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IntialExperimentStatus fills up ExperimentStatus Structure with intialValues
func (expStatus *ExperimentStatus) IntialExperimentStatus(experimentDetails *ExperimentDetails) {
	expStatus.Name = experimentDetails.JobName
	expStatus.Status = "Waiting for Job Creation"
	expStatus.Verdict = "Waiting"
	expStatus.LastUpdateTime = metav1.Now()
}

// AwaitingExperimentStatus fills up ExperimentStatus Structure with Running Status
func (expStatus *ExperimentStatus) AwaitingExperimentStatus(experimentDetails *ExperimentDetails) {
	expStatus.Name = experimentDetails.JobName
	expStatus.Status = "Running"
	expStatus.Verdict = "Awaited"
	expStatus.LastUpdateTime = metav1.Now()
}

// CompletedExperimentStatus fills up ExperimentStatus Structure with values chaosResult
func (expStatus *ExperimentStatus) CompletedExperimentStatus(chaosResult *v1alpha1.ChaosResult, experimentDetails *ExperimentDetails) {
	//var currExpStatus v1alpha1.ExperimentStatuses
	expStatus.Name = experimentDetails.JobName
	expStatus.Status = "Execution Successful"
	expStatus.LastUpdateTime = metav1.Now()
	expStatus.Verdict = chaosResult.Spec.ExperimentStatus.Verdict
	//return currExpStatus
}

// NotFoundExperimentStatus initilize experiment struct using the following values.
func (expStatus *ExperimentStatus) NotFoundExperimentStatus(experimentDetails *ExperimentDetails) {
	expStatus.Name = experimentDetails.JobName
	expStatus.Status = "ChaosExperiment Not Found"
	expStatus.Verdict = "Failed"
	expStatus.LastUpdateTime = metav1.Now()
}
