package utils

import (
	"github.com/pkg/errors"
)

// ExperimentNotFoundPatchEngine patches the chaosEngine when ChaosExperiment is not Found
func (engineDetails EngineDetails) ExperimentNotFoundPatchEngine(experiment *ExperimentDetails, clients ClientSets) error {

	var expStatus ExperimentStatus
	expStatus.NotFoundExperimentStatus(experiment.Name, engineDetails.Name)
	if err := expStatus.PatchChaosEngineStatus(engineDetails, clients); err != nil {
		return errors.Errorf("unable to Patch ChaosEngine with Status, error: %v", err)
	}
	return nil
}
