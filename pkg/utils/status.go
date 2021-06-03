package utils

import (
	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InitialExperimentStatus fills up ExperimentStatus Structure with InitialValues
func (expStatus *ExperimentStatus) InitialExperimentStatus(expName, engineName string) {
	expStatus.Name = expName
	expStatus.Runner = engineName + "-runner"
	expStatus.ExpPod = "Yet to be launched"
	expStatus.Status = v1alpha1.ExperimentStatusWaiting
	expStatus.Verdict = "N/A"
	expStatus.LastUpdateTime = metav1.Now()
}

// AwaitedExperimentStatus fills up ExperimentStatus Structure with Running Status
func (expStatus *ExperimentStatus) AwaitedExperimentStatus(expName, engineName, experimentPodName string) {
	expStatus.Name = expName
	expStatus.Runner = engineName + "-runner"
	expStatus.ExpPod = experimentPodName
	expStatus.Status = v1alpha1.ExperimentStatusRunning
	expStatus.Verdict = "Awaited"
	expStatus.LastUpdateTime = metav1.Now()
}

// CompletedExperimentStatus fills up ExperimentStatus Structure with values chaosResult
func (expStatus *ExperimentStatus) CompletedExperimentStatus(chaosResult *v1alpha1.ChaosResult, engineName, experimentPodName string) {
	//var currExpStatus v1alpha1.ExperimentStatuses
	expStatus.Name = chaosResult.Spec.ExperimentName
	expStatus.Runner = engineName + "-runner"
	expStatus.ExpPod = experimentPodName
	expStatus.Status = v1alpha1.ExperimentStatusCompleted
	expStatus.LastUpdateTime = metav1.Now()
	expStatus.Verdict = string(chaosResult.Status.ExperimentStatus.Verdict)
}

// NotFoundExperimentStatus initilize experiment struct using the following values.
func (expStatus *ExperimentStatus) NotFoundExperimentStatus(expName, engineName string) {
	expStatus.Name = expName
	expStatus.Runner = engineName + "-runner"
	expStatus.ExpPod = "N/A"
	expStatus.Status = v1alpha1.ExperimentStatusNotFound
	expStatus.Verdict = "Fail"
	expStatus.LastUpdateTime = metav1.Now()
}

// SkippedExperimentStatus fills up  ExperimentStatus Structure with skipped value
func (expStatus *ExperimentStatus) SkippedExperimentStatus(expName, engineName string) {
	expStatus.Name = expName
	expStatus.Runner = engineName + "-runner"
	expStatus.ExpPod = "N/A"
	expStatus.Status = v1alpha1.ExperimentSkipped
	expStatus.Verdict = "Fail"
	expStatus.LastUpdateTime = metav1.Now()
}
