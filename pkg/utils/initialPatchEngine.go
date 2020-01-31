package utils

import (
	"fmt"
	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
)

// ExperimentStatus is wrapper for v1alpha1.ExperimentStatuses
type ExperimentStatus v1alpha1.ExperimentStatuses

// InitialPatchEngine patches the chaosEngine with the initial ExperimentStatuses
func (expStatus *ExperimentStatus) InitialPatchEngine(engineDetails EngineDetails, clients ClientSets) {

	// TODO: check for the status before patching
	for range engineDetails.Experiments {
		Logger.WithString(fmt.Sprintf("Getting ChaosEngine for Patching")).WithVerbosity(1).Log()
		expEngine, err := engineDetails.GetChaosEngine(clients)
		if err != nil {
			Logger.WithNameSpace(engineDetails.AppNamespace).WithResourceName(engineDetails.Name).WithString(err.Error()).WithOperation("Get").WithVerbosity(1).WithResourceType("ChaosEngine").Log()
		}
		expEngine.Status.Experiments = append(expEngine.Status.Experiments, v1alpha1.ExperimentStatuses(*expStatus))
		//log.Info("Patching Engine")
		_, updateErr := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engineDetails.AppNamespace).Update(expEngine)
		if updateErr != nil {
			Logger.WithNameSpace(engineDetails.AppNamespace).WithResourceName(engineDetails.Name).WithString(err.Error()).WithOperation("Update").WithVerbosity(1).WithResourceType("ChaosEngine").Log()
		}
	}
}
