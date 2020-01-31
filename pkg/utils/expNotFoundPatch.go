package utils

import (
	"fmt"
)

// ExperimentNotFoundPatchEngine patches the chaosEngine when ChaosExperiment is not Found
func (engineDetails EngineDetails) ExperimentNotFoundPatchEngine(experiment *ExperimentDetails, clients ClientSets) {

	var expStatus ExperimentStatus

	Logger.WithString(fmt.Sprintf("Creating Not Found Experiment Status for ChaosEngine")).WithVerbosity(1).Log()
	expStatus.NotFoundExperimentStatus(experiment)

	if err := expStatus.PatchChaosEngineStatus(engineDetails, clients); err != nil {
		Logger.WithNameSpace(engineDetails.AppNamespace).WithResourceName(engineDetails.Name).WithString(err.Error()).WithOperation("Patch").WithVerbosity(1).WithResourceType("ChaosEngine").Log()
	}
}
