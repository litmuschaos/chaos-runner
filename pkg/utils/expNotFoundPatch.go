package utils

import (
	log "github.com/sirupsen/logrus"
)

// ExperimentNotFoundPatchEngine patches the chaosEngine when ChaosExperiment is not Found
func (engineDetails EngineDetails) ExperimentNotFoundPatchEngine(experiment *ExperimentDetails, clients ClientSets) {

	var expStatus ExperimentStatus
	expStatus.NotFoundExperimentStatus(experiment)
	if err := expStatus.PatchChaosEngineStatus(engineDetails, clients); err != nil {
		log.Infof("Unable to Patch ChaosEngine with Status, error: %v", err)
	}
}
