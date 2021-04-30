package utils

import (
	"github.com/pkg/errors"
)

// PatchResources function patches chaos Experiment Job with different Kubernetes Resources.
func (expDetails *ExperimentDetails) PatchResources(engineDetails EngineDetails, clients ClientSets) error {
	// Patch ConfigMaps to ChaosExperiment Job
	if err := expDetails.PatchConfigMaps(clients, engineDetails); err != nil {
		return errors.Errorf("unable to patch ConfigMaps to Chaos Experiment, error: %v", err)
	}

	// Patch Secrets to ChaosExperiment Job
	if err := expDetails.PatchSecrets(clients, engineDetails); err != nil {
		return errors.Errorf("unable to patch Secrets to Chaos Experiment, error: %v", err)
	}
	// Patch HostFileVolumes to ChaosExperiment Job
	if err := expDetails.PatchHostFileVolumes(clients, engineDetails); err != nil {
		return errors.Errorf("unable to patch hostFileVolumes to Chaos Experiment, error: %v", err)
	}
	return nil
}
