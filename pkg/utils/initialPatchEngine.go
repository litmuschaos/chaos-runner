package utils

import (
	log "github.com/sirupsen/logrus"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
)

// ExperimentStatus is wrapper for v1alpha1.ExperimentStatuses
type ExperimentStatus v1alpha1.ExperimentStatuses

// InitialPatchEngine patches the chaosEngine with the initial ExperimentStatuses
func (expStatus *ExperimentStatus) InitialPatchEngine(engineDetails EngineDetails, clients ClientSets) {

	// TODO: check for the status before patching
	for range engineDetails.Experiments {
		expEngine, err := engineDetails.GetChaosEngine(clients)
		if err != nil {
			log.Infof("Couldn't Get ChaosEngine: %v, wouldn't be able to update Status in ChaosEngine", err)
		}
		expEngine.Status.Experiments = append(expEngine.Status.Experiments, v1alpha1.ExperimentStatuses(*expStatus))
		//log.Info("Patching Engine")
		_, updateErr := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engineDetails.AppNamespace).Update(expEngine)
		if updateErr != nil {
			log.Infof("Couldn't Update ChaosEngine: %v, wouldn't be able to update Status in ChaosEngine", updateErr)
		}
	}
}
