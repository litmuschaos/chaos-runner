package utils

import (
	"k8s.io/klog"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
)

// ExperimentStatus is wrapper for v1alpha1.ExperimentStatuses
type ExperimentStatus v1alpha1.ExperimentStatuses

// InitialPatchEngine patches the chaosEngine with the initial ExperimentStatuses
func (expStatus *ExperimentStatus) InitialPatchEngine(engineDetails EngineDetails, clients ClientSets) {

	// TODO: check for the status before patching
	for range engineDetails.Experiments {

		klog.V(1).Infof("getting ChaosEngine for Patching")
		expEngine, err := engineDetails.GetChaosEngine(clients)
		if err != nil {
			klog.V(0).Infof("Unable to get ChaosEngine for Intial Patching")
			klog.V(1).Infof("Couldn't Get ChaosEngine: %v, wouldn't be able to update Status in ChaosEngine, due to error: &v", engineDetails.Name, err)
		}
		expEngine.Status.Experiments = append(expEngine.Status.Experiments, v1alpha1.ExperimentStatuses(*expStatus))
		//log.Info("Patching Engine")
		_, updateErr := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engineDetails.AppNamespace).Update(expEngine)
		if updateErr != nil {
			klog.V(0).Infof("Unable to update ChaosEngine for Intial Patching")
			klog.V(1).Infof("Couldn't Update ChaosEngine: %v, wouldn't be able to update Status in ChaosEngine", updateErr)
		}
	}
}
