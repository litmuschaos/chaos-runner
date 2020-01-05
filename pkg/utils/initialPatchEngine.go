package utils

import (
	log "github.com/sirupsen/logrus"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ExperimentStatus v1alpha1.ExperimentStatuses

// IntialExperimentStatus fills up ExperimentStatus Structure with intialValues
func (expStatus *ExperimentStatus) IntialExperimentStatus(experimentDetails *ExperimentDetails) {
	expStatus.Name = experimentDetails.JobName
	expStatus.Status = "Waiting for Job Creation"
	expStatus.Verdict = "Waiting"
	expStatus.LastUpdateTime = metav1.Now()
}

// InitialPatchEngine patches the chaosEngine with the initial ExperimentStatuses
func (expStatus *ExperimentStatus) InitialPatchEngine(engineDetails EngineDetails, clients ClientSets) {

	for i := range engineDetails.Experiments {
		log.Info("Initial Patch for Experiment : ", engineDetails.Experiments[i])
		expEngine, err := engineDetails.GetChaosEngine(clients)
		if err != nil {
			log.Infoln("Could'nt Get ChaosEngine : ", err)
		}
		expEngine.Status.Experiments = append(expEngine.Status.Experiments, v1alpha1.ExperimentStatuses(*expStatus))
		log.Info("Patching Engine")
		_, updateErr := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engineDetails.AppNamespace).Update(expEngine)
		if updateErr != nil {
			log.Infoln("Unable to Patch Engine, Update Error : ", updateErr)
		}
	}
}
