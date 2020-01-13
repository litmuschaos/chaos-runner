package utils

import (
	"k8s.io/klog"
)

// ExperimentNotFoundPatchEngine patches the chaosEngine when ChaosExperiment is not Found
func (engineDetails EngineDetails) ExperimentNotFoundPatchEngine(experiment *ExperimentDetails, clients ClientSets) {

	var expStatus ExperimentStatus

	klog.V(1).Infof("Creating Not Found Experiment Status")
	expStatus.NotFoundExperimentStatus(experiment)

	if err := expStatus.PatchChaosEngineStatus(engineDetails, clients); err != nil {
		klog.V(0).Infof("Unable to Patch ChaosEngine Status")
		klog.V(1).Infof("Unable to Patch ChaosEngine Status, error: %v", err)
	}
}
