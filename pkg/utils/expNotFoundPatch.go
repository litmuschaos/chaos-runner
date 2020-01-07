package utils

import ()

// ExperimentNotFoundPatchEngine patches the chaosEngine when ChaosExperiment is not Found
func (engineDetails EngineDetails) ExperimentNotFoundPatchEngine(experiment *ExperimentDetails, clients ClientSets) {

	var expStatus ExperimentStatus
	expStatus.NotFoundExperimentStatus(experiment)
	expStatus.PatchChaosEngineStatus(engineDetails, clients)
}
