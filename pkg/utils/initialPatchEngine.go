package utils

import (
	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	"github.com/pkg/errors"
)

// ExperimentStatus is wrapper for v1alpha1.ExperimentStatuses
type ExperimentStatus v1alpha1.ExperimentStatuses

// InitialPatchEngine patches the chaosEngine with the initial ExperimentStatuses
func (expStatus *ExperimentStatus) InitialPatchEngine(engineDetails EngineDetails, clients ClientSets) error {

	// TODO: check for the status before patching
	for range engineDetails.Experiments {
		expEngine, err := engineDetails.GetChaosEngine(clients)
		if err != nil {
			return errors.Wrapf(err, "Unable to get ChaosEngine, due to error: %v", err)
		}
		expEngine.Status.Experiments = append(expEngine.Status.Experiments, v1alpha1.ExperimentStatuses(*expStatus))
		_, updateErr := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engineDetails.AppNamespace).Update(expEngine)
		if updateErr != nil {
			return errors.Wrapf(err, "Unable to update ChaosEngine in namespace: %v, due to error: %v", engineDetails.AppNamespace, err)
		}
	}
	return nil
}
