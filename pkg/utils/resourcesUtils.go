package utils

import (
	"github.com/pkg/errors"
)

// PatchResources function patches chaos Experiment Job with different Kubernetes Resources.
func (expDetails *ExperimentDetails) PatchResources(engineDetails EngineDetails, clients ClientSets) error {
	// Patch ConfigMaps to ChaosExperiment Job
	if err := expDetails.PatchConfigMaps(clients, engineDetails); err != nil {
		return errors.Wrapf(err, "Unable to patch ConfigMaps to Chaos Experiment, due to error: %v", err)
	}

	// Patch Secrets to ChaosExperiment Job
	if err := expDetails.PatchSecrets(clients, engineDetails); err != nil {
		return errors.Wrapf(err, "Unable to patch Secrets to Chaos Experiment, due to error: %v", err)
	}
	return nil
}
